//go:build integration

package kafka

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"log"
	"os"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"go.uber.org/zap"
)

var (
	testKafkaContainer *kafka.KafkaContainer
	testBrokers        []string
	testTopic          string
	testAdminClient    sarama.ClusterAdmin
	testContext        context.Context
)

func TestMain(m *testing.M) {
	var err error
	testContext = context.Background()

	testKafkaContainer, err = kafka.Run(testContext, "confluentinc/confluent-local:7.5.0")
	if err != nil {
		log.Fatalf("Error al iniciar contenedor Kafka: %v", err)
	}

	testBrokers, err = testKafkaContainer.Brokers(testContext)
	if err != nil {
		log.Fatalf("Error al obtener brokers: %v", err)
	}

	testAdminClient, err = sarama.NewClusterAdmin(testBrokers, sarama.NewConfig())
	if err != nil {
		log.Fatalf("Error al crear cliente admin: %v", err)
	}

	testTopic = "test-download-request"
	err = testAdminClient.CreateTopic(testTopic, &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1,
	}, false)
	if err != nil {
		log.Fatalf("Error al crear tópico: %v", err)
	}

	code := m.Run()

	if err := testAdminClient.Close(); err != nil {
		log.Printf("Error al cerrar cliente admin: %v", err)
	}
	if err := testKafkaContainer.Terminate(testContext); err != nil {
		log.Printf("Error al terminar contenedor: %v", err)
	}

	os.Exit(code)
}

func createTestConfig(brokers []string) *config.Config {
	return &config.Config{
		QueueConfig: config.QueueConfig{
			KafkaConfig: config.KafkaConfig{
				Brokers: brokers,
				Topics: &config.KafkaTopics{
					BotDownloadRequest: "test-download-request",
				},
				TLS: shared.TLSConfig{
					Enabled: false,
				},
			},
		},
	}
}

func createTestLogger(t *testing.T) logging.Logger {
	zapLogger, err := logging.NewDevelopmentLogger()
	require.NoError(t, err)
	return zapLogger
}

func setupConsumer(t *testing.T, brokers []string) sarama.Consumer {
	cfgKafka := sarama.NewConfig()
	cfgKafka.Consumer.Return.Errors = true
	cfgKafka.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumer(brokers, cfgKafka)
	require.NoError(t, err, "Error al crear consumidor de prueba")

	return consumer
}

func TestProducerKafka_PublishSongRequest(t *testing.T) {
	cfg := createTestConfig(testBrokers)
	testLogger := createTestLogger(t)

	producer, err := NewProducerKafka(cfg, testLogger)
	require.NoError(t, err, "Error al crear el productor Kafka")
	defer func() {
		if err := producer.Close(); err != nil {
			t.Fatalf("Error al cerrar productor: %v", err)
		}
	}()

	consumer := setupConsumer(t, testBrokers)
	defer func() {
		if err := consumer.Close(); err != nil {
			t.Fatalf("Error al cerrar consumidor: %v", err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition(testTopic, 0, sarama.OffsetOldest)
	require.NoError(t, err, "Error al crear el consumidor de partición")
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			t.Fatalf("Error al cerrar consumidor de partición: %v", err)
		}
	}()

	testMessage := &entity.SongRequestMessage{
		InteractionID: "test-interaction-123",
		UserID:        "user123",
		Song:          "Despacito",
		ProviderType:  "youtube",
		Timestamp:     time.Now(),
	}

	err = producer.PublishSongRequest(testContext, testMessage)
	require.NoError(t, err, "Error al publicar el mensaje")

	select {
	case msg := <-partitionConsumer.Messages():
		assert.Equal(t, "test-interaction-123", string(msg.Key))
		assert.Contains(t, string(msg.Value), "test-interaction-123")
		assert.Contains(t, string(msg.Value), "user123")
		assert.Contains(t, string(msg.Value), "Despacito")
		testLogger.Info("Mensaje recibido correctamente",
			zap.String("topic", msg.Topic),
			zap.ByteString("value", msg.Value))
	case err := <-partitionConsumer.Errors():
		t.Fatalf("Error al consumir mensaje: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout esperando el mensaje")
	}
}

func TestProducerKafka_PublishSongRequest_ContextCancellation(t *testing.T) {
	cfg := createTestConfig(testBrokers)
	testLogger := createTestLogger(t)

	producer, err := NewProducerKafka(cfg, testLogger)
	require.NoError(t, err, "Error al crear el productor Kafka")
	defer func() {
		if err := producer.Close(); err != nil {
			t.Fatalf("Error al cerrar productor: %v", err)
		}
	}()

	cancelCtx, cancel := context.WithCancel(testContext)
	testMessage := &entity.SongRequestMessage{
		InteractionID: "test-canceled-interaction",
		UserID:        "user123",
		ProviderType:  "youtube",
		Timestamp:     time.Now(),
	}

	cancel()

	err = producer.PublishSongRequest(cancelCtx, testMessage)
	assert.Error(t, err, "Se esperaba un error debido a la cancelación del contexto")
	assert.Contains(t, err.Error(), "contexto cancelado")
}
