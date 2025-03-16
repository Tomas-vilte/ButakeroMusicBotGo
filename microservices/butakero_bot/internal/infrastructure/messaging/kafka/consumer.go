package kafka

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/IBM/sarama"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
)

type ConsumerClient interface {
	Partitions(topic string) ([]int32, error)
	ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error)
	Close() error
}

type KafkaConsumer struct {
	consumer    sarama.Consumer
	brokers     []string
	topic       string
	logger      logging.Logger
	messageChan chan *entity.MessageQueue
	errorChan   chan error
	wg          sync.WaitGroup
}

func NewKafkaConsumer(config KafkaConfig, logger logging.Logger) (*KafkaConsumer, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	logger.Info("Iniciando configuración del consumidor Kafka",
		zap.Strings("brokers", config.Brokers),
		zap.String("topic", config.Topic))

	if config.TLS {
		logger.Info("Configurando conexión TLS")
		tlsConfig, err := NewTLSConfig(config.CertFile, config.KeyFile, config.CACertFile)
		if err != nil {
			logger.Error("Error al crear configuración TLS", zap.Error(err))
			return nil, fmt.Errorf("error creando configuración TLS: %w", err)
		}
		saramaConfig.Net.TLS.Enable = true
		saramaConfig.Net.TLS.Config = tlsConfig
	}

	consumer, err := sarama.NewConsumer(config.Brokers, saramaConfig)
	if err != nil {
		logger.Error("Error al crear consumidor Kafka",
			zap.Error(err),
			zap.Strings("brokers", config.Brokers))
		return nil, fmt.Errorf("error creando configuración Kafka consumer: %w", err)
	}

	return &KafkaConsumer{
		consumer:    consumer,
		brokers:     config.Brokers,
		topic:       config.Topic,
		logger:      logger,
		messageChan: make(chan *entity.MessageQueue),
		errorChan:   make(chan error),
	}, nil
}

func (k *KafkaConsumer) ConsumeMessages(ctx context.Context, offset int64) error {
	k.logger.Info("Iniciando consumo de mensajes", zap.String("topic", k.topic))

	exists, err := k.TopicExists(k.topic)
	if err != nil {
		k.logger.Error("Error al verificar si el topic existe", zap.Error(err))
		return err
	}

	if !exists {
		k.logger.Error("El topic no existe", zap.String("topic", k.topic))
		return fmt.Errorf("el topic '%s' no existe", k.topic)
	}

	partitionList, err := k.consumer.Partitions(k.topic)
	if err != nil {
		k.logger.Error("Error al obtener particiones", zap.String("topic", k.topic), zap.Error(err))
		return fmt.Errorf("error al obtener las particiones: %w", err)
	}

	k.logger.Info("Particiones obtenidas", zap.Int("amount", len(partitionList)))

	for _, partition := range partitionList {
		pc, err := k.consumer.ConsumePartition(k.topic, partition, offset)
		if err != nil {
			k.logger.Error("Error al crear consumidor de partición", zap.Int32("partition", partition), zap.Error(err))
			return fmt.Errorf("error al crear la particion del consumidor: %w", err)
		}

		k.wg.Add(1)
		go k.consumePartition(ctx, pc, partition)
	}

	go func() {
		k.wg.Wait()
		close(k.messageChan)
		close(k.errorChan)
	}()

	return nil
}

func (k *KafkaConsumer) consumePartition(ctx context.Context, pc sarama.PartitionConsumer, partition int32) {
	defer k.wg.Done()
	defer func() {
		if err := pc.Close(); err != nil {
			k.logger.Error("Error al cerrar partición del consumidor", zap.Int32("partition", partition), zap.Error(err))
		}
	}()

	k.logger.Debug("Consumiendo mensajes de la partición", zap.Int32("partition", partition))

	for {
		select {
		case msg := <-pc.Messages():
			k.logger.Info("Mensaje recibido en la partición",
				zap.Int32("partition", partition),
				zap.Int64("offset", msg.Offset))
			k.handleMessage(msg)
		case err := <-pc.Errors():
			k.logger.Error("Error consumiendo mensajes", zap.Int32("partition", partition), zap.Error(err))
			k.errorChan <- err
		case <-ctx.Done():
			k.logger.Info("Consumidor cerrado", zap.Int32("partition", partition))
			return
		}
	}
}

func (k *KafkaConsumer) handleMessage(msg *sarama.ConsumerMessage) {
	k.logger.Debug("Mensaje recibido", zap.Int64("offset", msg.Offset))

	var statusMessage entity.MessageQueue
	if err := json.Unmarshal(msg.Value, &statusMessage); err != nil {
		k.logger.Error("Error al deserializar mensaje", zap.Error(err), zap.ByteString("contenido_mensaje", msg.Value))
		k.errorChan <- err
		return
	}

	k.logger.Debug("Mensaje procesado",
		zap.String("status", statusMessage.Status),
		zap.String("videoID", statusMessage.VideoID))

	if statusMessage.Status == "success" {
		k.logger.Info("Mensaje procesado exitosamente", zap.String("status", statusMessage.Status))
		k.messageChan <- &statusMessage
	} else {
		k.logger.Warn("Mensaje con estado no exitoso", zap.String("status", statusMessage.Status))
	}
}

func (k *KafkaConsumer) TopicExists(topic string) (bool, error) {
	topics, err := k.consumer.Topics()
	if err != nil {
		k.logger.Error("Error al obtener la lista de topics", zap.Error(err))
		return false, fmt.Errorf("error al obtener la lista de topics: %w", err)
	}

	for _, t := range topics {
		if t == topic {
			k.logger.Info("El topic existe", zap.String("topic", topic))
			return true, nil
		}
	}
	k.logger.Warn("El topic no existe", zap.String("topic", topic))
	return false, nil
}

func (k *KafkaConsumer) GetMessagesChannel() <-chan *entity.MessageQueue {
	return k.messageChan
}

func (k *KafkaConsumer) GetErrorChannel() <-chan error {
	return k.errorChan
}

func (k *KafkaConsumer) Close() error {
	k.logger.Info("Cerrando consumidor Kafka")
	return k.consumer.Close()
}

func NewTLSConfig(certFile, keyFile, caCertFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("error cargando par de claves: %w", err)
	}

	caCert, err := os.ReadFile(caCertFile)
	if err != nil {
		return nil, fmt.Errorf("error leyendo certificado CA: %w", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	return tlsConfig, nil
}
