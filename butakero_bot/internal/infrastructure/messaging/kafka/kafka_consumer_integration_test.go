//go:build integration

package kafka

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/kafka"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
)

func TestIntegrationKafkaConsumer(t *testing.T) {
	ctx := context.Background()
	kafkaContainer, err := kafka.Run(ctx,
		"confluentinc/confluent-local:7.5.0",
	)
	require.NoError(t, err)
	defer func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Fatal("Error al eliminar contenedor")
		}
	}()

	brokers, err := kafkaContainer.Brokers(ctx)
	require.NoError(t, err)

	topic := "test-topic"

	prodConfig := sarama.NewConfig()
	prodConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(brokers, prodConfig)
	require.NoError(t, err)
	defer func() {
		if err := producer.Close(); err != nil {
			t.Fatal("Error al cerrar productor")
		}
	}()

	successJSON := `{
		"user_id": "user_123",
		"interaction_id": "interaction_123",
        "video_id": "19f6c66f-26f3-4ccf-bfc7-967449a95ad4",
        "status": "success",
        "message": "Procesamiento exitoso",
        "platform_metadata": {
            "title": "The Emptiness Machine (Official Music Video) - Linkin Park",
            "duration_ms": 2345,
            "url": "https://youtube.com/watch?v=SRXH9AbT280",
            "thumbnail_url": "https://i.ytimg.com/vi/SRXH9AbT280/default.jpg",
            "platform": "youtube"
        },
        "file_data": {
            "file_path": "audio/The Emptiness Machine (Official Music Video) - Linkin Park.dca",
            "file_size": "1.44MB",
            "file_type": "audio/dca"
        },
        "success": true
    }`

	errorJSON := `{
        "video_id": "DFswyIQyrl8",
        "status": "error",
        "message": "Error en descarga: io: read/write on closed pipe",
        "platform_metadata": {
            "title": "Ke Personajes - My Immortal (Video Oficial)",
            "duration_ms": 259000,
            "url": "https://youtube.com/watch?v=DFswyIQyrl8",
            "thumbnail_url": "https://i.ytimg.com/vi/DFswyIQyrl8/default.jpg",
            "platform": "youtube"
        },
		"file_data": null,
        "success": false
    }`

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(successJSON),
	})
	require.NoError(t, err)
	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(errorJSON),
	})
	require.NoError(t, err)

	configKafka := ConfigKafka{
		Brokers: brokers,
		Topic:   topic,
		Offset:  -2,
		TLS:     shared.TLSConfig{},
	}

	logger, err := logging.NewDevelopmentLogger()
	require.NoError(t, err)
	consumer, err := NewKafkaConsumer(configKafka, logger)
	require.NoError(t, err)

	ctxConsume, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err = consumer.SubscribeToDownloadEvents(ctxConsume)
	require.NoError(t, err)

	select {
	case received := <-consumer.DownloadEventsChannel():
		assert.Equal(t, "success", received.Status, "Se esperaba un mensaje con estado 'success'")
		assert.Equal(t, "Procesamiento exitoso", received.Message, "El mensaje debe coincidir")
	case <-time.After(5 * time.Second):
		t.Fatal("No se recibió el mensaje de éxito esperado")
	}
}
