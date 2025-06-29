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
)

type ConsumerKafka struct {
	saramaConsumer sarama.Consumer
	logger         logger.Logger
	cfg            *config.Config
}

func NewConsumerKafka(cfg *config.Config, logger logger.Logger) (ports.MessageConsumer, error) {
	logger.Info("Inicializando consumidor Kafka",
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
	cfgKafka.Consumer.Return.Errors = true

	if cfg.Messaging.Kafka.EnableTLS {
		cfgKafka.Net.TLS.Enable = true
		cfgKafka.Net.TLS.Config = tlsConfig
	} else {
		cfgKafka.Net.TLS.Enable = false
	}

	logger.Debug("Creando consumidor Sarama")
	c, err := sarama.NewConsumer(cfg.Messaging.Kafka.Brokers, cfgKafka)
	if err != nil {
		logger.Error("Error al crear consumidor Sarama", zap.Error(err))
		return nil, errors.ErrKafkaConnectionFailed.Wrap(err)
	}

	consumerKafka := &ConsumerKafka{
		saramaConsumer: c,
		logger:         logger,
		cfg:            cfg,
	}

	logger.Debug("Verificando existencia del tópico",
		zap.String("topic", cfg.Messaging.Kafka.Topics.BotDownloadRequests))
	if err := consumerKafka.EnsureTopicExists(cfg.Messaging.Kafka.Topics.BotDownloadRequests); err != nil {
		logger.Error("Error al verificar tópico", zap.Error(err))
		return nil, errors.ErrKafkaTopicCreation.Wrap(err)
	}

	logger.Info("Consumidor Kafka inicializado correctamente")
	return consumerKafka, nil
}

func (c *ConsumerKafka) GetRequestsChannel(ctx context.Context) (<-chan *model.MediaRequest, error) {
	log := c.logger.With(
		zap.String("component", "ConsumerKafka"),
		zap.String("method", "GetRequestsChannel"),
		zap.String("topic", c.cfg.Messaging.Kafka.Topics.BotDownloadRequests),
	)

	log.Info("Iniciando consumo de mensajes")

	log.Debug("Consumiendo partición",
		zap.String("topic", c.cfg.Messaging.Kafka.Topics.BotDownloadRequests),
		zap.Int32("partition", 0),
		zap.Int64("offset", sarama.OffsetNewest))

	partitionConsumer, err := c.saramaConsumer.ConsumePartition(c.cfg.Messaging.Kafka.Topics.BotDownloadRequests, 0, sarama.OffsetNewest)
	if err != nil {
		log.Error("Error al consumir la partición", zap.Error(err))
		return nil, errors.ErrKafkaMessageConsume.Wrap(err)
	}

	out := make(chan *model.MediaRequest)
	log.Debug("Canal de solicitudes creado")

	go c.consumeLoop(ctx, partitionConsumer, out)
	log.Info("Rutina de consumo iniciada")

	return out, nil
}

func (c *ConsumerKafka) consumeLoop(ctx context.Context, pc sarama.PartitionConsumer, out chan<- *model.MediaRequest) {
	log := c.logger.With(
		zap.String("component", "ConsumerKafka"),
		zap.String("method", "consumeLoop"),
		zap.String("topic", c.cfg.Messaging.Kafka.Topics.BotDownloadRequests),
	)

	log.Info("Loop de consumo iniciado")

	defer func() {
		log.Debug("Cerrando canal de salida")
		close(out)

		log.Debug("Cerrando consumidor de partición")
		if err := pc.Close(); err != nil {
			log.Error("Error al cerrar la partición", zap.Error(err))
		}
		log.Info("Loop de consumo finalizado")
	}()

	msgCount := 0
	for {
		select {
		case msg := <-pc.Messages():
			offset := msg.Offset
			msgCount++

			log.Debug("Mensaje recibido",
				zap.Int64("offset", offset),
				zap.Int("count", msgCount),
				zap.Int("payload_size", len(msg.Value)))

			var request model.MediaRequest
			if err := json.Unmarshal(msg.Value, &request); err != nil {
				log.Error("Error deserializando mensaje",
					zap.Error(errors.ErrKafkaMessageConsume.Wrap(err)),
					zap.Int64("offset", offset),
					zap.ByteString("payload", msg.Value))
				continue
			}

			log.Info("Mensaje procesado correctamente",
				zap.String("request_id", request.RequestID),
				zap.String("user_id", request.UserID),
				zap.String("provider_type", request.ProviderType))

			out <- &request

		case err := <-pc.Errors():
			log.Error("Error en consumidor", zap.Error(errors.ErrKafkaMessageConsume.Wrap(err)))
		case <-ctx.Done():
			log.Info("Contexto cancelado, finalizando loop de consumo")
			return
		}
	}
}

func (c *ConsumerKafka) Close() error {
	log := c.logger.With(
		zap.String("component", "ConsumerKafka"),
		zap.String("method", "Close"),
	)

	log.Info("Cerrando consumidor Kafka")

	if err := c.saramaConsumer.Close(); err != nil {
		log.Error("Error cerrando el consumidor", zap.Error(err))
		return errors.ErrKafkaConnectionFailed.Wrap(err)
	}

	log.Info("Consumidor Kafka cerrado correctamente")
	return nil
}

func (c *ConsumerKafka) EnsureTopicExists(topic string) error {
	log := c.logger.With(
		zap.String("component", "ConsumerKafka"),
		zap.String("method", "EnsureTopicExists"),
		zap.String("topic", topic),
	)

	cfgKafka := sarama.NewConfig()

	if c.cfg.Messaging.Kafka.EnableTLS {
		tlsConfig, err := utils.NewTLSConfig(&utils.TLSConfig{
			CaFile:   c.cfg.Messaging.Kafka.CaFile,
			CertFile: c.cfg.Messaging.Kafka.CertFile,
			KeyFile:  c.cfg.Messaging.Kafka.KeyFile,
		})
		if err != nil {
			return err
		}
		cfgKafka.Net.TLS.Enable = true
		cfgKafka.Net.TLS.Config = tlsConfig
	}

	log.Debug("Creando administrador de clúster Kafka")
	admin, err := sarama.NewClusterAdmin(c.cfg.Messaging.Kafka.Brokers, cfgKafka)
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
	}

	return nil
}
