package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue"
	serviceSqs "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue/sqs"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQSServiceIntegration(t *testing.T) {
	t.Parallel()

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

	t.Run("TestMessageLifecycle", func(t *testing.T) {
		t.Parallel()

		ctx := context.Background()
		uniqueID := uuid.New().String()
		message := queue.Message{
			ID:      fmt.Sprintf("test-msg-%s", uniqueID),
			Content: fmt.Sprintf("Test Message Content %s", uniqueID),
		}

		// Enviar mensaje
		err = service.SendMessage(ctx, message)
		require.NoError(t, err, "Error al enviar mensaje")

		// Función auxiliar para buscar nuestro mensaje específico
		findOurMessage := func(messages []queue.Message) *queue.Message {
			for _, msg := range messages {
				if msg.ID == message.ID {
					return &msg
				}
			}
			return nil
		}

		// Intentar recibir el mensaje con reintentos más cortos
		var receivedMessage *queue.Message
		for i := 0; i < 3; i++ {
			messages, err := service.ReceiveMessage(ctx)
			require.NoError(t, err, "Error al recibir mensajes")

			if msg := findOurMessage(messages); msg != nil {
				receivedMessage = msg
				break
			}
			time.Sleep(time.Second)
		}

		// Verificar que el mensaje se recibió correctamente
		require.NotNil(t, receivedMessage, "No se pudo encontrar nuestro mensaje específico")
		assert.Equal(t, message.ID, receivedMessage.ID, "ID del mensaje no coincide")
		assert.Equal(t, message.Content, receivedMessage.Content, "Contenido del mensaje no coincide")
		assert.NotEmpty(t, receivedMessage.ReceiptHandle, "ReceiptHandle está vacío")

		// Eliminar el mensaje
		err = service.DeleteMessage(ctx, receivedMessage.ReceiptHandle)
		require.NoError(t, err, "Error al eliminar mensaje")

		// Verificar que el mensaje ya no está disponible
		time.Sleep(2 * time.Second) // Reducido el tiempo de espera

		var messageStillExists bool
		for i := 0; i < 2; i++ { // Reducido el número de intentos
			messages, err := service.ReceiveMessage(ctx)
			require.NoError(t, err, "Error al verificar eliminación del mensaje")

			if msg := findOurMessage(messages); msg != nil {
				messageStillExists = true
				break
			}
			time.Sleep(time.Second)
		}

		assert.False(t, messageStillExists, "El mensaje no debería estar disponible después de ser eliminado")
	})
}
