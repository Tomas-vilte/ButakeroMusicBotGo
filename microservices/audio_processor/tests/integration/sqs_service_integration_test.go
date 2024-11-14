package integration

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	serviceSqs "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue/sqs"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func cleanupMessages(ctx context.Context, service *serviceSqs.SQSService) error {
	for {
		messages, err := service.ReceiveMessage(ctx)
		if err != nil {
			return err
		}
		if len(messages) == 0 {
			break
		}

		for _, msg := range messages {
			if err := service.DeleteMessage(ctx, msg.ReceiptHandle); err != nil {
				return err
			}
		}
	}
	return nil
}

func setupTestEnvironment(t *testing.T) (*serviceSqs.SQSService, *config.Config) {
	if testing.Short() {
		t.Skip("Saltando test de integración en modo corto")
	}

	cfg := &config.Config{
		AWS: config.AWSConfig{
			Region: os.Getenv("REGION"),
			Credentials: &config.CredentialsConfig{
				AccessKey: os.Getenv("ACCESS_KEY"),
				SecretKey: os.Getenv("SECRET_KEY"),
			},
		},
		Messaging: config.MessagingConfig{
			SQS: &config.SQSConfig{
				QueueURL: os.Getenv("SQS_QUEUE_URL"),
			},
		},
	}

	if cfg.Messaging.SQS.QueueURL == "" || cfg.AWS.Region == "" {
		t.Fatal("SQS_QUEUE_URL y REGION no fueron seteados para los tests de integración")
	}

	log, err := logger.NewZapLogger()
	require.NoError(t, err)

	service, err := serviceSqs.NewSQSService(cfg, log)
	require.NoError(t, err)

	err = cleanupMessages(context.Background(), service)
	require.NoError(t, err)

	return service, cfg
}

func createTestMessage() model.Message {
	return model.Message{
		ID:      uuid.New().String(),
		Content: "Test message " + time.Now().Format(time.RFC3339),
		Status: model.Status{
			ID:       uuid.New().String(),
			Status:   "pending",
			Message:  "Test status message",
			Success:  false,
			Attempts: 0,
		},
	}
}

func TestSQSServiceIntegration(t *testing.T) {
	service, _ := setupTestEnvironment(t)
	ctx := context.Background()

	defer func() {
		err := cleanupMessages(ctx, service)
		assert.NoError(t, err, "Error durante la limpieza final de mensajes")
	}()

	t.Run("SendMessage Success", func(t *testing.T) {
		message := createTestMessage()
		err := service.SendMessage(ctx, message)
		assert.NoError(t, err)

		messages, err := service.ReceiveMessage(ctx)
		assert.NoError(t, err)
		for _, msg := range messages {
			err = service.DeleteMessage(ctx, msg.ReceiptHandle)
			assert.NoError(t, err)
		}
	})

	t.Run("ReceiveMessage Success", func(t *testing.T) {
		sentMessage := createTestMessage()
		err := service.SendMessage(ctx, sentMessage)
		assert.NoError(t, err)

		messages, err := service.ReceiveMessage(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, messages)

		err = service.DeleteMessage(ctx, messages[0].ReceiptHandle)
		assert.NoError(t, err)
	})

	t.Run("DeleteMessage Success", func(t *testing.T) {
		sentMessage := createTestMessage()
		err := service.SendMessage(ctx, sentMessage)
		require.NoError(t, err)

		messages, err := service.ReceiveMessage(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, messages)

		err = service.DeleteMessage(ctx, messages[0].ReceiptHandle)
		assert.NoError(t, err)
	})

	t.Run("SendMessage InvalidContext", func(t *testing.T) {
		message := createTestMessage()
		ctx, cancel := context.WithCancel(ctx)
		cancel()

		err := service.SendMessage(ctx, message)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("DeleteMessage EmptyReceiptHandle", func(t *testing.T) {
		emptyReceiptHandler := ""
		err := service.DeleteMessage(ctx, emptyReceiptHandler)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "receipt handle no puede estar vacio")
	})

	t.Run("Concurrent Message Processing", func(t *testing.T) {
		numMessages := 5
		messages := make([]model.Message, numMessages)
		for i := 0; i < numMessages; i++ {
			messages[i] = createTestMessage()
		}

		errChan := make(chan error, numMessages)
		for _, msg := range messages {
			go func(m model.Message) {
				errChan <- service.SendMessage(ctx, m)
			}(msg)
		}

		for i := 0; i < numMessages; i++ {
			assert.NoError(t, <-errChan)
		}

		receivedMessages, err := service.ReceiveMessage(ctx)
		assert.NoError(t, err)
		for _, msg := range receivedMessages {
			err := service.DeleteMessage(ctx, msg.ReceiptHandle)
			assert.NoError(t, err)
		}
	})
}

func TestSQSService_ErrorCases(t *testing.T) {
	t.Run("NewSQSService nil config", func(t *testing.T) {
		log, _ := logger.NewZapLogger()
		service, err := serviceSqs.NewSQSService(nil, log)
		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "config no puede ser nil")
	})

	t.Run("NewSQSService nil logger", func(t *testing.T) {
		cfg := &config.Config{}
		service, err := serviceSqs.NewSQSService(cfg, nil)
		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "logger no puede ser nil")
	})
}
