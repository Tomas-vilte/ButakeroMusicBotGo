package kafka

import (
	"context"
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
		"status": {
			"id": "19f6c66f-26f3-4ccf-bfc7-967449a95ad4",
			"sk": "SRXH9AbT280",
			"status": "success",
			"message": "Procesamiento exitoso",
			"metadata": {
				"id": "63f48016-78cd-4387-99b9-c38af46e8e90",
				"video_id": "SRXH9AbT280",
				"title": "The Emptiness Machine (Official Music Video) - Linkin Park",
				"duration": "PT3M21S",
				"url_youtube": "https://youtube.com/watch?v=SRXH9AbT280",
				"thumbnail": "https://i.ytimg.com/vi/SRXH9AbT280/default.jpg",
				"platform": "Youtube"
			},
			"file_data": {
				"file_path": "audio/The Emptiness Machine (Official Music Video) - Linkin Park.dca",
				"file_size": "1.44MB",
				"file_type": "audio/dca",
				"public_url": "file://data/audio-files/audio/The Emptiness Machine (Official Music Video) - Linkin Park.dca"
			},
			"processing_date": "2024-12-24T05:39:58Z",
			"success": true,
			"attempts": 1,
			"failures": 0
		}
	}`

	errorJSON := `{
		"status": {
			"id": "959326a1-53db-4810-9fc8-b17275122158",
			"sk": "DFswyIQyrl8",
			"status": "error",
			"message": "Error en descarga: io: read/write on closed pipe",
			"metadata": {
				"id": "873f0521-f808-4721-b2c1-5e63a782b7cf",
				"video_id": "DFswyIQyrl8",
				"title": "Ke Personajes - My Immortal (Video Oficial)",
				"duration": "PT4M19S",
				"url_youtube": "https://youtube.com/watch?v=DFswyIQyrl8",
				"thumbnail": "https://i.ytimg.com/vi/DFswyIQyrl8/default.jpg",
				"platform": "Youtube"
			},
			"file_data": null,
			"processing_date": "2025-02-10T14:23:14Z",
			"success": false,
			"attempts": 8,
			"failures": 8
		}
	}`

	// Publicar ambos mensajes en el topic real
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

	configKafka := KafkaConfig{
		Brokers: brokers,
		Topic:   topic,
		TLS:     false,
	}

	logger, err := logging.NewZapLogger()
	require.NoError(t, err)
	consumer, err := NewKafkaConsumer(configKafka, logger)
	require.NoError(t, err)

	ctxConsume, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	err = consumer.ConsumeMessages(ctxConsume, sarama.OffsetOldest)
	require.NoError(t, err)

	select {
	case received := <-consumer.GetMessagesChannel():
		assert.Equal(t, "success", received.Status.Status, "Se esperaba un mensaje con estado 'success'")
		assert.Equal(t, "Procesamiento exitoso", received.Status.Message, "El mensaje debe coincidir")
	case <-time.After(5 * time.Second):
		t.Fatal("No se recibió el mensaje de éxito esperado")
	}
}
