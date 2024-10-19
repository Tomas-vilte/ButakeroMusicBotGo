package integration

import (
	"context"
	"encoding/json"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue"
	serviceSqs "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue/sqs"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestSQSServiceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Saltando test de integración en modo corto")
	}

	cfg := config.Config{
		QueueURL:  os.Getenv("SQS_QUEUE_URL"),
		Region:    os.Getenv("REGION"),
		AccessKey: os.Getenv("ACCESS_KEY"),
		SecretKey: os.Getenv("SECRET_KEY"),
	}

	if cfg.QueueURL == "" || cfg.Region == "" {
		t.Fatal("SQS_QUEUE_URL y REGION no fueron seteados para los tests de integración")
	}

	log, err := logger.NewZapLogger()
	require.NoError(t, err)

	service, err := serviceSqs.NewSQSService(cfg, log)
	require.NoError(t, err)

	t.Run("SendAndReceiveMessage", func(t *testing.T) {
		// arrange
		message := queue.Message{
			ID:      "Integration-test-id",
			Content: "Integration Test Message",
		}

		// act - SendMessage
		err = service.SendMessage(context.Background(), message)
		require.NoError(t, err)

		var receivedMessage *types.Message
		for i := 0; i < 3; i++ {
			receivedMessage, err = service.ReceiveMessage(context.Background())
			require.NoError(t, err)

			if receivedMessage != nil {
				break
			}
			time.Sleep(time.Second * 5)
		}

		// assert - verificamos que el mensaje se recibió correctamente
		require.NotNil(t, receivedMessage, "No se recibio ningun mensaje")

		var receivedMessageBody queue.Message
		err = json.Unmarshal([]byte(*receivedMessage.Body), &receivedMessageBody)
		require.NoError(t, err)

		assert.Equal(t, message.ID, receivedMessageBody.ID)
		assert.Equal(t, message.Content, receivedMessageBody.Content)

		// act DeleteMessage
		err = service.DeleteMessage(context.Background(), *receivedMessage.ReceiptHandle)
		require.NoError(t, err)

		deletedMessage, err := service.ReceiveMessage(context.Background())
		require.NoError(t, err)

		assert.Nil(t, deletedMessage, "El mensaje no deberia estar disponible despues de eliminarlo")
	})

	t.Run("ReceiveAndDeleteMessage", func(t *testing.T) {
		message := queue.Message{
			ID:      "Integration-test-id",
			Content: "Integration Test Message",
		}

		err = service.SendMessage(context.Background(), message)
		require.NoError(t, err)

		var receivedMessage *types.Message
		for i := 0; i < 3; i++ {
			receivedMessage, err = service.ReceiveMessage(context.Background())
			if err == nil && receivedMessage != nil {
				break
			}
			time.Sleep(time.Second * 2)
		}

		require.NoError(t, err)
		require.NotNil(t, receivedMessage, "No se recibio ningun mensaje")

		var receivedMessageBody queue.Message
		err = json.Unmarshal([]byte(*receivedMessage.Body), &receivedMessageBody)
		require.NoError(t, err)

		assert.Equal(t, message.ID, receivedMessageBody.ID)
		assert.Equal(t, message.Content, receivedMessageBody.Content)

		err = service.DeleteMessage(context.Background(), *receivedMessage.ReceiptHandle)
		require.NoError(t, err)

		deletedMessage, err := service.ReceiveMessage(context.Background())
		require.NoError(t, err)
		assert.Nil(t, deletedMessage, "El mensaje no deberia estar disponible despues de eliminarlo")
	})
}
