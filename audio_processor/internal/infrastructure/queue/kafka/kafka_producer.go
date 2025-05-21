package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/utils"
	"go.uber.org/zap"
	"time"
)

type ProducerKafka struct {
	saramaProducer sarama.SyncProducer
	logger         logger.Logger
	config         *config.Config
}

func NewProducerKafka(cfg *config.Config, logger logger.Logger) (ports.MessageProducer, error) {
	logger.Info("Inicializando productor Kafka",
		zap.Strings("brokers", cfg.Messaging.Kafka.Brokers),
		zap.Bool("tls_enabled", cfg.Messaging.Kafka.EnableTLS))

	var tlsConfig *tls.Config
	var err error

	if cfg.Messaging.Kafka.EnableTLS {
		logger.Debug("Configurando TLS para Kafka")
		tlsConfig, err = utils.NewTLSConfig(&utils.TLSConfig{
			CaFile:   cfg.Messaging.Kafka.CaFile,
			CertFile: cfg.Messaging.Kafka.CertFile,
			KeyFile:  cfg.Messaging.Kafka.KeyFile,
		})
		if err != nil {
			logger.Error("Error en configuración TLS", zap.Error(err))
			return nil, errors.ErrKafkaTLSConfig.Wrap(err)
		}
	}

	cfgKafka := sarama.NewConfig()
	cfgKafka.Producer.Return.Successes = true

	logger.Debug("Configurando cliente Kafka",
		zap.Bool("return_successes", cfgKafka.Producer.Return.Successes))

	if cfg.Messaging.Kafka.EnableTLS {
		cfgKafka.Net.TLS.Enable = true
		cfgKafka.Net.TLS.Config = tlsConfig
	} else {
		cfgKafka.Net.TLS.Enable = false
	}

	logger.Debug("Creando productor Sarama")
	p, err := sarama.NewSyncProducer(cfg.Messaging.Kafka.Brokers, cfgKafka)
	if err != nil {
		logger.Error("Error al crear productor Sarama", zap.Error(err))
		return nil, errors.ErrKafkaConnectionFailed.Wrap(err)
	}

	producerKafka := &ProducerKafka{
		saramaProducer: p,
		logger:         logger,
		config:         cfg,
	}

	logger.Debug("Verificando existencia del tópico",
		zap.String("topic", cfg.Messaging.Kafka.Topics.BotDownloadStatus))
	if err := producerKafka.EnsureTopicExists(cfg.Messaging.Kafka.Topics.BotDownloadStatus); err != nil {
		logger.Error("Error al verificar tópico", zap.Error(err))
		return nil, errors.ErrKafkaTopicCreation.Wrap(err)
	}

	logger.Info("Productor Kafka inicializado correctamente")
	return producerKafka, nil
}

func (p *ProducerKafka) Publish(ctx context.Context, msg *model.MediaProcessingMessage) error {
	log := p.logger.With(
		zap.String("component", "ProducerKafka"),
		zap.String("method", "Publish"),
		zap.String("topic", p.config.Messaging.Kafka.Topics.BotDownloadStatus),
	)

	select {
	case <-ctx.Done():
		return errors.ErrKafkaMessagePublish.Wrap(ctx.Err()).WithMessage("contexto cancelado antes de publicar mensaje")
	default:
	}

	log.Debug("Serializando mensaje para publicar")
	payload, err := json.Marshal(msg)
	if err != nil {
		log.Error("Error al serializar el mensaje", zap.Error(err))
		return errors.ErrKafkaMessagePublish.Wrap(err)
	}

	kafkaMsg := &sarama.ProducerMessage{
		Topic:     p.config.Messaging.Kafka.Topics.BotDownloadStatus,
		Key:       sarama.StringEncoder(msg.VideoID),
		Value:     sarama.ByteEncoder(payload),
		Timestamp: time.Now(),
	}

	done := make(chan error, 1)
	go func() {
		partition, offset, err := p.saramaProducer.SendMessage(kafkaMsg)
		if err != nil {
			done <- err
			return
		}
		log.Info("Mensaje publicado exitosamente",
			zap.Int32("partition", partition),
			zap.Int64("offset", offset),
			zap.String("video_id", msg.VideoID),
			zap.String("status", msg.Status),
			zap.Bool("success", msg.Success))
		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return errors.ErrKafkaMessagePublish.Wrap(ctx.Err()).WithMessage("contexto cancelado antes de recibir respuesta")
	}
}

func (p *ProducerKafka) Close() error {
	log := p.logger.With(
		zap.String("component", "ProducerKafka"),
		zap.String("method", "Close"),
	)

	log.Info("Cerrando productor Kafka")

	if err := p.saramaProducer.Close(); err != nil {
		log.Error("Error al cerrar el productor", zap.Error(err))
		return errors.ErrKafkaConnectionFailed.Wrap(err)
	}

	log.Info("Productor Kafka cerrado correctamente")
	return nil
}

func (p *ProducerKafka) EnsureTopicExists(topic string) error {
	log := p.logger.With(
		zap.String("component", "ProducerKafka"),
		zap.String("method", "EnsureTopicExists"),
		zap.String("topic", topic),
	)

	cfgKafka := sarama.NewConfig()

	if p.config.Messaging.Kafka.EnableTLS {
		tlsConfig, err := utils.NewTLSConfig(&utils.TLSConfig{
			CaFile:   p.config.Messaging.Kafka.CaFile,
			CertFile: p.config.Messaging.Kafka.CertFile,
			KeyFile:  p.config.Messaging.Kafka.KeyFile,
		})
		if err != nil {
			return errors.ErrKafkaTLSConfig.Wrap(err)
		}
		cfgKafka.Net.TLS.Enable = true
		cfgKafka.Net.TLS.Config = tlsConfig
	}

	log.Debug("Creando administrador de clúster Kafka")
	admin, err := sarama.NewClusterAdmin(p.config.Messaging.Kafka.Brokers, cfgKafka)
	if err != nil {
		log.Error("Error al crear el administrador de Kafka", zap.Error(err))
		return errors.ErrKafkaAdminClient.Wrap(err)
	}
	defer func() {
		log.Debug("Cerrando administrador de clúster")
		if err := admin.Close(); err != nil {
			log.Error("Error al cerrar el administrador de Kafka", zap.Error(err))
		}
	}()

	log.Debug("Listando tópicos existentes")
	topics, err := admin.ListTopics()
	if err != nil {
		log.Error("Error al listar los tópicos", zap.Error(err))
		return errors.ErrKafkaAdminClient.Wrap(err)
	}

	if _, exists := topics[topic]; !exists {
		log.Info("El tópico no existe, creándolo...", zap.String("topic", topic))

		topicDetail := &sarama.TopicDetail{
			NumPartitions:     1,
			ReplicationFactor: 1,
		}
		log.Debug("Detalles del tópico a crear",
			zap.Int32("partitions", topicDetail.NumPartitions),
			zap.Int16("replication_factor", topicDetail.ReplicationFactor))

		err := admin.CreateTopic(topic, topicDetail, false)
		if err != nil {
			log.Error("Error al crear el tópico", zap.Error(err))
			return errors.ErrKafkaTopicCreation.Wrap(err)
		}
		log.Info("Tópico creado exitosamente", zap.String("topic", topic))
	} else {
		log.Info("El tópico ya existe", zap.String("topic", topic))

		log.Debug("Obteniendo detalles del tópico existente")
		metadata, err := admin.DescribeTopics([]string{topic})
		if err != nil {
			log.Warn("No se pudieron obtener detalles del tópico", zap.Error(err))
		} else if len(metadata) > 0 {
			log.Debug("Detalles del tópico",
				zap.String("topic", metadata[0].Name),
				zap.Int("partition_count", len(metadata[0].Partitions)))
		}
	}
	return nil
}
