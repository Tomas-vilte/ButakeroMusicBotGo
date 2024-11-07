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

func setupTestEnvironment(t *testing.T) (*serviceSqs.SQSService, *config.Config) {
	if testing.Short() {
		t.Skip("Saltando test de integración en modo corto")
	}

	cfg := &config.Config{
		AWS: &config.AWSConfig{
			Region: os.Getenv("REGION"),
			Credentials: config.CredentialsConfig{
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

	t.Run("SendMessage Success", func(t *testing.T) {
		// arrange
		message := createTestMessage()

		// act
		err := service.SendMessage(ctx, message)

		// assert
		assert.NoError(t, err)
	})

	t.Run("ReceiveMessage Success", func(t *testing.T) {
		// arrange
		sentMessage := createTestMessage()
		err := service.SendMessage(ctx, sentMessage)
		assert.NoError(t, err)

		messages, err := service.ReceiveMessage(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, messages)

		// act
		err = service.DeleteMessage(ctx, messages[0].ReceiptHandle)

		// assert
		assert.NoError(t, err)
	})

	t.Run("DeleteMessage Success", func(t *testing.T) {
		// Arrange
		sentMessage := createTestMessage()
		err := service.SendMessage(ctx, sentMessage)
		require.NoError(t, err)

		messages, err := service.ReceiveMessage(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, messages)

		// Act
		err = service.DeleteMessage(ctx, messages[0].ReceiptHandle)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("SendMessage InvalidContext", func(t *testing.T) {
		// arrange
		message := createTestMessage()
		ctx, cancel := context.WithCancel(ctx)
		cancel()

		// act
		err := service.SendMessage(ctx, message)

		// assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")

	})

	t.Run("DeleteMessage EmptyReceiptHandle", func(t *testing.T) {
		// arrange
		emptyReceiptHandler := ""

		// act
		err := service.DeleteMessage(ctx, emptyReceiptHandler)

		// assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "receipt handle no puede estar vacio")
	})

	t.Run("Concurrent Message Processing", func(t *testing.T) {
		// arrange
		numMessages := 5
		messages := make([]model.Message, numMessages)
		for i := 0; i < numMessages; i++ {
			messages[i] = createTestMessage()
		}

		// act & assert: Send Messages concurrent
		errChan := make(chan error, numMessages)
		for _, msg := range messages {
			go func(m model.Message) {
				errChan <- service.SendMessage(ctx, m)
			}(msg)
		}

		for i := 0; i < numMessages; i++ {
			assert.NoError(t, <-errChan)
		}

		// act & assert: Receive message
		receivedMessages, err := service.ReceiveMessage(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, receivedMessages)

		for _, msg := range receivedMessages {
			err := service.DeleteMessage(ctx, msg.ReceiptHandle)
			assert.NoError(t, err)
		}
	})
}

func TestSQSService_ErrorCases(t *testing.T) {
	t.Run("NewSQSService nil config", func(t *testing.T) {
		// arrange
		log, _ := logger.NewZapLogger()

		// act
		service, err := serviceSqs.NewSQSService(nil, log)

		// assert
		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "config no puede ser nil")
	})

	t.Run("NewSQSService nil logger", func(t *testing.T) {
		// arrange
		cfg := &config.Config{}

		// act
		service, err := serviceSqs.NewSQSService(cfg, nil)

		// assert
		assert.Error(t, err)
		assert.Nil(t, service)
		assert.Contains(t, err.Error(), "logger no puede ser nil")
	})
}
