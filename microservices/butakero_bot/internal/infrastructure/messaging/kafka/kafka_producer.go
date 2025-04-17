package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model/queue"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
	"time"
)

type ProducerKafka struct {
	producer sarama.SyncProducer
	cfg      *config.Config
	logger   logging.Logger
}

func NewProducerKafka(cfg *config.Config, logger logging.Logger) (ports.SongDownloadRequestPublisher, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Return.Successes = true

	logger = logger.With(
		zap.String("component", "kafka_consumer"),
		zap.Strings("brokers", cfg.QueueConfig.KafkaConfig.Brokers),
		zap.String("topic", cfg.QueueConfig.KafkaConfig.Topics.BotDownloadRequest),
	)

	logger.Info("Iniciando configuración del producer Kafka")

	if cfg.QueueConfig.KafkaConfig.TLS.Enabled {
		logger.Info("Configurando conexión TLS")
		tlsConfig, err := shared.ConfigureTLS(shared.TLSConfig{
			Enabled:  cfg.QueueConfig.KafkaConfig.TLS.Enabled,
			CAFile:   cfg.QueueConfig.KafkaConfig.TLS.CAFile,
			CertFile: cfg.QueueConfig.KafkaConfig.TLS.CertFile,
			KeyFile:  cfg.QueueConfig.KafkaConfig.TLS.KeyFile,
		})
		if err != nil {
			logger.Error("Error al crear configuración TLS", zap.Error(err))
			return nil, fmt.Errorf("error creando configuración TLS: %w", err)
		}
		saramaConfig.Net.TLS.Enable = true
		saramaConfig.Net.TLS.Config = tlsConfig
		logger.Debug("Configuración TLS aplicada")
	} else {
		logger.Info("TLS no está habilitado")
	}

	producer, err := sarama.NewSyncProducer(cfg.QueueConfig.KafkaConfig.Brokers, saramaConfig)
	if err != nil {
		logger.Error("Error al crear el productor de Sarama", zap.Error(err))
		return nil, fmt.Errorf("error al crear el productor de Sarama: %w", err)
	}

	return &ProducerKafka{
		producer: producer,
		logger:   logger,
		cfg:      cfg,
	}, nil
}

func (p *ProducerKafka) PublishDownloadRequest(ctx context.Context, message *queue.DownloadRequestMessage) error {
	log := p.logger.With(
		zap.String("component", "ProducerKafka"),
		zap.String("method", "PublishDownloadRequest"),
		zap.String("topic", p.cfg.QueueConfig.KafkaConfig.Topics.BotDownloadRequest),
		zap.String("request_id", message.RequestID),
	)

	select {
	case <-ctx.Done():
		return fmt.Errorf("contexto cancelado antes de publicar: %w", ctx.Err())
	default:
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Error("Error al serializar el mensaje", zap.Error(err))
		return fmt.Errorf("error al serializar el mensaje: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic:     p.cfg.QueueConfig.KafkaConfig.Topics.BotDownloadRequest,
		Key:       sarama.StringEncoder(message.RequestID),
		Value:     sarama.ByteEncoder(jsonData),
		Timestamp: time.Now(),
	}

	done := make(chan error, 1)
	go func() {
		partition, offset, err := p.producer.SendMessage(msg)
		if err != nil {
			log.Error("Error al enviar el mensaje", zap.Error(err))
			done <- fmt.Errorf("error al enviar el mensaje: %w", err)
			return
		}
		log.Info("Mensaje publicado exitosamente",
			zap.Int32("partition", partition),
			zap.Int64("offset", offset),
			zap.String("request_id", message.RequestID))
		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operación cancelada mientras se publicaba: %w", ctx.Err())
	}
}

func (p *ProducerKafka) ClosePublisher() error {
	logger := p.logger.With(
		zap.String("component", "ProducerKafka"),
		zap.String("method", "ClosePublisher"))
	logger.Info("Cerrando producer Kafka")
	return p.producer.Close()
}
