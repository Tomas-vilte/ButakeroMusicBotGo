package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/utils"
	"go.uber.org/zap"
)

type KafkaService struct {
	Config   *config.Config
	Producer sarama.SyncProducer
	Consumer sarama.Consumer
	Log      logger.Logger
}

func NewKafkaService(cfgApplication *config.Config, log logger.Logger) (*KafkaService, error) {
	var tlsConfig *tls.Config
	var err error

	if cfgApplication.Messaging.Kafka.EnableTLS {
		tlsConfig, err = utils.NewTLSConfig(&utils.TLSConfig{
			CaFile:   cfgApplication.Messaging.Kafka.CaFile,
			CertFile: cfgApplication.Messaging.Kafka.CertFile,
			KeyFile:  cfgApplication.Messaging.Kafka.KeyFile,
		})
		if err != nil {
			return nil, errors.ErrCodeDBConnectionFailed.WithMessage(fmt.Sprintf("Error configurando conexion de TLS de Kafka: %v", err))
		}
	}

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Consumer.Return.Errors = true

	if cfgApplication.Messaging.Kafka.EnableTLS {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = tlsConfig
	} else {
		cfg.Net.TLS.Enable = false
	}

	producer, err := sarama.NewSyncProducer(cfgApplication.Messaging.Kafka.Brokers, cfg)
	if err != nil {
		return nil, errors.ErrCodeDBConnectionFailed.WithMessage(fmt.Sprintf("Error al crear el productor de Kafka: %v", err))
	}

	consumer, err := sarama.NewConsumer(cfgApplication.Messaging.Kafka.Brokers, cfg)
	if err != nil {
		return nil, errors.ErrCodeDBConnectionFailed.WithMessage(fmt.Sprintf("Error al crear el consumidor de Kafka: %v", err))
	}

	kafkaService := &KafkaService{
		Config:   cfgApplication,
		Producer: producer,
		Consumer: consumer,
		Log:      log,
	}

	if err := kafkaService.EnsureTopicExists(cfgApplication.Messaging.Kafka.Topic); err != nil {
		return nil, errors.ErrCodeDBConnectionFailed.WithMessage(fmt.Sprintf("Error al asegurar que el topic existe: %v", err))
	}
	return kafkaService, nil
}

func (k *KafkaService) SendMessage(ctx context.Context, message *model.MediaProcessingMessage) error {
	log := k.Log.With(
		zap.String("component", "KafkaService"),
		zap.String("method", "SendMessage"),
	)
	body, err := json.Marshal(message)
	if err != nil {
		log.Error("Error al serializar el mensaje", zap.Error(err))
		return errors.ErrPublishMessageFailed.WithMessage(fmt.Sprintf("error al serializar mensaje: %v", err))
	}

	msg := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(message.VideoID),
		Topic: k.Config.Messaging.Kafka.Topic,
		Value: sarama.StringEncoder(body),
	}

	partition, offset, err := k.Producer.SendMessage(msg)
	if err != nil {
		log.Error("Error al enviar el mensaje", zap.Error(err))
		return errors.ErrPublishMessageFailed.WithMessage(fmt.Sprintf("error al enviar mensaje a kafka: %v", err))
	}

	log.Info("Mensaje enviado con éxito",
		zap.Int32("partition", partition),
		zap.Int64("offset", offset))

	return nil
}

func (k *KafkaService) ReceiveMessage(ctx context.Context) ([]model.MediaProcessingMessage, error) {
	log := k.Log.With(
		zap.String("component", "KafkaService"),
		zap.String("method", "ReceiveMessage"),
	)
	partitionConsumer, err := k.Consumer.ConsumePartition(k.Config.Messaging.Kafka.Topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Error("Error al crear la partición del consumidor", zap.Error(err))
		return nil, errors.ErrPublishMessageFailed.WithMessage(fmt.Sprintf("error al crear la particion del consumidor: %v", err))
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			log.Error("Error al cerrar la partición del consumidor", zap.Error(err))
		}
	}()

	var messages []model.MediaProcessingMessage

	select {
	case msg := <-partitionConsumer.Messages():
		log.Info("Mensaje recibido desde Kafka", zap.String("message_id", string(msg.Key)))
		var message model.MediaProcessingMessage
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			log.Error("Error al deserializar el mensaje", zap.Error(err))
			return nil, errors.ErrPublishMessageFailed.WithMessage(fmt.Sprintf("error al deserializar mensaje: %v", err))
		}
		messages = append(messages, message)
	case <-ctx.Done():
		log.Error("Contexto cancelado durante la recepción de mensajes", zap.Error(ctx.Err()))
		return messages, errors.ErrPublishMessageFailed.WithMessage(fmt.Sprintf("contexto cancelado durante la recepción de mensajes: %v", ctx.Err()))
	}

	log.Info("Mensajes recibidos exitosamente", zap.Int("count", len(messages)))
	return messages, nil
}

func (k *KafkaService) DeleteMessage(_ context.Context, receiptHandle string) error {
	log := k.Log.With(
		zap.String("component", "KafkaService"),
		zap.String("method", "DeleteMessage"),
		zap.String("receipt_handle", receiptHandle),
	)
	log.Info("Se llamó a DeleteMessage, pero Kafka no admite la eliminación de mensajes individuales")
	return nil
}

func (k *KafkaService) Close() error {
	log := k.Log.With(
		zap.String("component", "KafkaService"),
		zap.String("method", "Close"),
	)

	log.Info("Cerrando el productor de Kafka")
	return k.Producer.Close()
}

func (k *KafkaService) EnsureTopicExists(topic string) error {
	log := k.Log.With(
		zap.String("component", "KafkaService"),
		zap.String("method", "EnsureTopicExists"),
		zap.String("topic", topic),
	)

	admin, err := sarama.NewClusterAdmin(k.Config.Messaging.Kafka.Brokers, sarama.NewConfig())
	if err != nil {
		log.Error("Error al crear el administrador de Kafka", zap.Error(err))
		return errors.ErrCodeDBConnectionFailed.WithMessage(fmt.Sprintf("error al crear el administrador de Kafka: %v", err))
	}
	defer func() {
		if err := admin.Close(); err != nil {
			log.Error("Error al cerrar el administrador de Kafka", zap.Error(err))
		}
	}()

	topics, err := admin.ListTopics()
	if err != nil {
		log.Error("Error al listar los topics", zap.Error(err))
		return errors.ErrCodeDBConnectionFailed.WithMessage(fmt.Sprintf("error al listar los topics: %v", err))
	}

	if _, exists := topics[topic]; !exists {
		log.Info("El topic no existe, creándolo...", zap.String("topic", topic))
		err := admin.CreateTopic(topic, &sarama.TopicDetail{
			NumPartitions:     1,
			ReplicationFactor: 1,
		}, false)
		if err != nil {
			log.Error("Error al crear el topic", zap.Error(err))
			return errors.ErrCodeDBConnectionFailed.WithMessage(fmt.Sprintf("error al crear el topic: %v", err))
		}
		log.Info("Topic creado exitosamente", zap.String("topic", topic))
	} else {
		log.Info("El topic ya existe", zap.String("topic", topic))
	}

	return nil
}
