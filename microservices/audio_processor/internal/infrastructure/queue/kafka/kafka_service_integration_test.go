//go:build integration

package kafka

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	kafkaContainerService "github.com/testcontainers/testcontainers-go/modules/kafka"
	"testing"
	"time"
)

func TestKafkaIntegration_SendAndReceiveMessage(t *testing.T) {
	ctx := context.Background()

	kafkaContainer, err := kafkaContainerService.Run(ctx, "confluentinc/confluent-local:7.5.0")
	assert.NoError(t, err)

	defer func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	if kafkaContainer.IsRunning() != true {
		t.Fatalf("Kafka no esta corriendo")
	}

	brokers, err := kafkaContainer.Brokers(ctx)
	assert.NoError(t, err)

	cfg := &config.Config{
		Messaging: config.MessagingConfig{
			Kafka: &config.KafkaConfig{
				Brokers:   brokers,
				Topic:     "test-topic",
				EnableTLS: false,
			},
		},
	}

	log, err := logger.NewProductionLogger()
	assert.NoError(t, err)

	kafkaService, err := NewKafkaService(cfg, log)
	assert.NoError(t, err)

	message := &model.MediaProcessingMessage{
		VideoID: "test-video",
	}

	err = kafkaService.SendMessage(ctx, message)
	assert.NoError(t, err)

	receiveCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	receivedMessages, err := kafkaService.ReceiveMessage(receiveCtx)
	if err != nil {
		t.Fatalf("Error al recibir mensaje: %v", err)
	}

	assert.NotEmpty(t, receivedMessages, "Se esperaban mensajes, pero no se recibió ninguno")
	assert.Equal(t, message.VideoID, receivedMessages[0].VideoID, "El contenido recibido no coincide con el enviado")
	assert.Equal(t, message, receivedMessages[0], "Respuesta no coincide con el enviado")

}
