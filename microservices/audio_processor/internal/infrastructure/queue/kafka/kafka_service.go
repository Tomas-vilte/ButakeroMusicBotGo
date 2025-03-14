package kafka

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// KafkaService proporciona métodos para interactuar con Kafka para producir y consumir mensajes.
type KafkaService struct {
	Config   *config.Config
	Producer sarama.SyncProducer
	Consumer sarama.Consumer
	Log      logger.Logger
}

// NewKafkaService crea una nueva instancia de KafkaService.
// Inicializa el productor y el consumidor de Kafka en base a la configuración y el logger proporcionados.
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
			return nil, errors.Wrap(err, "Error configurando conexion de TLS de Kafka")
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
		return nil, err
	}

	consumer, err := sarama.NewConsumer(cfgApplication.Messaging.Kafka.Brokers, cfg)
	if err != nil {
		return nil, err
	}

	return &KafkaService{
		Config:   cfgApplication,
		Producer: producer,
		Consumer: consumer,
		Log:      log,
	}, nil
}

// SendMessage envía un mensaje al tema de Kafka especificado en la configuración.
// Serializa el mensaje a JSON y registra el resultado.
func (k *KafkaService) SendMessage(ctx context.Context, message *model.MediaProcessingMessage) error {
	log := k.Log.With(
		zap.String("component", "KafkaService"),
		zap.String("method", "SendMessage"),
	)
	body, err := json.Marshal(message)
	if err != nil {
		log.Error("Error al serializar el mensaje", zap.Error(err))
		return errors.Wrap(err, "error deserializar mensaje")
	}

	msg := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(message.VideoID),
		Topic: k.Config.Messaging.Kafka.Topic,
		Value: sarama.StringEncoder(body),
	}

	partition, offset, err := k.Producer.SendMessage(msg)
	if err != nil {
		log.Error("Error al enviar el mensaje", zap.Error(err))
		return errors.Wrap(err, "error al enviar mensaje a kafka")
	}

	log.Info("Mensaje enviado con éxito",
		zap.Int32("partition", partition),
		zap.Int64("offset", offset))

	return nil
}

// ReceiveMessage recibe mensajes del tema de Kafka especificado en la configuración.
// Deserializa los mensajes de JSON y registra el resultado.
func (k *KafkaService) ReceiveMessage(ctx context.Context) ([]model.MediaProcessingMessage, error) {
	log := k.Log.With(
		zap.String("component", "KafkaService"),
		zap.String("method", "ReceiveMessage"),
	)
	partitionConsumer, err := k.Consumer.ConsumePartition(k.Config.Messaging.Kafka.Topic, 0, sarama.OffsetOldest)
	if err != nil {
		log.Error("Error al crear la partición del consumidor", zap.Error(err))
		return nil, errors.Wrap(err, "error al crear la particion del consumidor")
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
			return nil, errors.Wrap(err, "error al deserializar mensaje")
		}
		messages = append(messages, message)
	case <-ctx.Done():
		log.Error("Contexto cancelado durante la recepción de mensajes", zap.Error(ctx.Err()))
		return messages, ctx.Err()
	}

	log.Info("Mensajes recibidos exitosamente", zap.Int("count", len(messages)))
	return messages, nil
}

// DeleteMessage registra un mensaje indicando que Kafka no admite la eliminación de mensajes individuales.
// Los mensajes se eliminan automáticamente según el período de retención.
func (k *KafkaService) DeleteMessage(ctx context.Context, receiptHandle string) error {
	log := k.Log.With(
		zap.String("component", "KafkaService"),
		zap.String("method", "DeleteMessage"),
		zap.String("receipt_handle", receiptHandle),
	)
	log.Info("Se llamó a DeleteMessage, pero Kafka no admite la eliminación de mensajes individuales")
	return nil
}

// Close cierra el productor de Kafka.
func (k *KafkaService) Close() error {
	log := k.Log.With(
		zap.String("component", "KafkaService"),
		zap.String("method", "Close"),
	)

	log.Info("Cerrando el productor de Kafka")
	return k.Producer.Close()
}
