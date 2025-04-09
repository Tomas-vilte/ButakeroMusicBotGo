//go:build integration

package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	kafkaContainerService "github.com/testcontainers/testcontainers-go/modules/kafka"
)

func setupKafkaContainer(t *testing.T) ([]string, func()) {
	ctx := context.Background()

	kafkaContainer, err := kafkaContainerService.Run(ctx, "confluentinc/confluent-local:7.5.0")
	require.NoError(t, err)

	brokers, err := kafkaContainer.Brokers(ctx)
	require.NoError(t, err)

	return brokers, func() {
		if err := kafkaContainer.Terminate(ctx); err != nil {
			t.Logf("Error terminating Kafka container: %v", err)
		}
	}
}

func TestKafkaIntegration_Producer(t *testing.T) {
	brokers, teardown := setupKafkaContainer(t)
	defer teardown()

	cfg := &config.Config{
		Messaging: config.MessagingConfig{
			Kafka: &config.KafkaConfig{
				Brokers:   brokers,
				Topics:    &config.KafkaTopics{BotDownloadRequests: "test-requests", BotDownloadStatus: "test-status"},
				EnableTLS: false,
			},
		},
	}

	log, err := logger.NewDevelopmentLogger()
	require.NoError(t, err)

	producer, err := NewProducerKafka(cfg, log)
	require.NoError(t, err)
	defer producer.Close()

	t.Run("Publish message to status topic", func(t *testing.T) {
		msg := &model.MediaProcessingMessage{
			VideoID: "test-video-1",
			Status:  "processing",
		}

		err = producer.Publish(context.Background(), msg)
		assert.NoError(t, err)
	})
}

func TestKafkaIntegration_TopicCreation(t *testing.T) {
	brokers, teardown := setupKafkaContainer(t)
	defer teardown()

	cfg := &config.Config{
		Messaging: config.MessagingConfig{
			Kafka: &config.KafkaConfig{
				Brokers:   brokers,
				Topics:    &config.KafkaTopics{BotDownloadRequests: "new-test-topic", BotDownloadStatus: "new-test-status"},
				EnableTLS: false,
			},
		},
	}

	log, err := logger.NewDevelopmentLogger()
	require.NoError(t, err)

	t.Run("Producer creates topic if not exists", func(t *testing.T) {
		producer, err := NewProducerKafka(cfg, log)
		require.NoError(t, err)
		defer producer.Close()

		// Verify topic was created by trying to publish
		msg := &model.MediaProcessingMessage{
			VideoID: "test-video-3",
			Status:  "processing",
		}
		err = producer.Publish(context.Background(), msg)
		assert.NoError(t, err)
	})

	t.Run("Consumer creates topic if not exists", func(t *testing.T) {
		consumer, err := NewConsumerKafka(cfg, log)
		require.NoError(t, err)
		defer consumer.Close()

		// Verify topic was created by trying to consume
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = consumer.GetRequestsChannel(ctx)
		assert.NoError(t, err)
	})
}

func TestKafkaIntegration_ErrorCases(t *testing.T) {
	brokers, teardown := setupKafkaContainer(t)
	defer teardown()

	log, err := logger.NewDevelopmentLogger()
	require.NoError(t, err)

	t.Run("Invalid broker addresses", func(t *testing.T) {
		invalidCfg := &config.Config{
			Messaging: config.MessagingConfig{
				Kafka: &config.KafkaConfig{
					Brokers:   []string{"invalid:9092"},
					Topics:    &config.KafkaTopics{BotDownloadRequests: "test", BotDownloadStatus: "test"},
					EnableTLS: false,
				},
			},
		}

		_, err := NewProducerKafka(invalidCfg, log)
		assert.Error(t, err)

		_, err = NewConsumerKafka(invalidCfg, log)
		assert.Error(t, err)
	})

	t.Run("TLS configuration fails with invalid certs", func(t *testing.T) {
		invalidTlsCfg := &config.Config{
			Messaging: config.MessagingConfig{
				Kafka: &config.KafkaConfig{
					Brokers:   brokers,
					Topics:    &config.KafkaTopics{BotDownloadRequests: "test", BotDownloadStatus: "test"},
					EnableTLS: true,
					CaFile:    "invalid.ca",
					CertFile:  "invalid.crt",
					KeyFile:   "invalid.key",
				},
			},
		}

		_, err := NewProducerKafka(invalidTlsCfg, log)
		assert.Error(t, err)

		_, err = NewConsumerKafka(invalidTlsCfg, log)
		assert.Error(t, err)
	})
}
