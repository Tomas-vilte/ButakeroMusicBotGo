package integration

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue"
	serviceSqs "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue/sqs"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
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

		// Esperar un poco para asegurarse de que el mensaje esté disponible
		time.Sleep(time.Second * 5)

		var receivedMessages []queue.Message
		for i := 0; i < 3; i++ {
			receivedMessages, err = service.ReceiveMessage(context.Background())
			require.NoError(t, err)

			if len(receivedMessages) > 0 {
				break
			}
			time.Sleep(time.Second * 5)
		}

		// assert - verificamos que el mensaje se recibió correctamente
		require.NotEmpty(t, receivedMessages, "No se recibió ningún mensaje")

		receivedMessage := receivedMessages[0]
		assert.Equal(t, message.ID, receivedMessage.ID)
		assert.Equal(t, message.Content, receivedMessage.Content)
		assert.NotEmpty(t, receivedMessage.ReceiptHandle, "ReceiptHandle no debe estar vacío")

		// act DeleteMessage
		err = service.DeleteMessage(context.Background(), receivedMessage.ReceiptHandle)
		require.NoError(t, err)

		// Esperar un poco más para asegurarse de que el mensaje se ha eliminado
		time.Sleep(time.Second * 10)

		// Intentar recibir mensajes nuevamente
		deletedMessages, err := service.ReceiveMessage(context.Background())
		require.NoError(t, err)

		// Verificar que no se recibió el mensaje eliminado
		for _, msg := range deletedMessages {
			assert.NotEqual(t, message.ID, msg.ID, "El mensaje eliminado no debería estar disponible")
		}
	})

	t.Run("ReceiveAndDeleteMessage", func(t *testing.T) {
		// Enviar un mensaje para la prueba
		message := queue.Message{
			ID:      "Integration-test-id-2",
			Content: "Integration Test Message 2",
		}

		err := service.SendMessage(context.Background(), message)
		require.NoError(t, err, "Error al enviar el mensaje")

		receivedMessages, err := service.ReceiveMessage(context.Background())
		require.NoError(t, err, "Error al recibir mensajes de la cola")
		require.NotEmpty(t, receivedMessages, "No se recibió ningún mensaje")

		receivedMessage := receivedMessages[0]
		assert.Equal(t, message.ID, receivedMessage.ID, "El ID del mensaje no coincide")

		err = service.DeleteMessage(context.Background(), receivedMessage.ReceiptHandle)
		require.NoError(t, err, "Error al eliminar el mensaje de la cola")

		time.Sleep(6 * time.Second)

		emptyMessages, err := service.ReceiveMessage(context.Background())
		require.NoError(t, err, "Error al recibir mensajes de la cola luego de eliminar")
		assert.Empty(t, emptyMessages, "La cola debería estar vacía luego de eliminar el mensaje")
	})
}
