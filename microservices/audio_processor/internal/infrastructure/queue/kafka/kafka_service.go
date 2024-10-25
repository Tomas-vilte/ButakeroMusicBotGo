package kafka

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/port"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type KafkaService struct {
	Config   config.Config
	Producer sarama.SyncProducer
	Consumer sarama.Consumer
	Log      logger.Logger
}

func NewKafkaService(cfgApplication config.Config, log logger.Logger) (*KafkaService, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.Return.Successes = true
	cfg.Consumer.Return.Errors = true

	producer, err := sarama.NewSyncProducer(cfgApplication.Brokers, cfg)
	if err != nil {
		return nil, err
	}

	consumer, err := sarama.NewConsumer(cfgApplication.Brokers, cfg)
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

func (k *KafkaService) SendMessage(ctx context.Context, message port.Message) error {
	body, err := json.Marshal(message)
	if err != nil {
		return errors.Wrap(err, "error deserializar mensaje")
	}

	msg := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(message.ID),
		Topic: k.Config.Topic,
		Value: sarama.StringEncoder(body),
	}

	partition, offset, err := k.Producer.SendMessage(msg)
	if err != nil {
		return errors.Wrap(err, "error al enviar mensaje a kafka")
	}

	k.Log.Info("Mensaje enviado con exito",
		zap.String("messageID", message.ID),
		zap.Int32("partition", partition),
		zap.Int64("offset", offset))

	return nil
}

func (k *KafkaService) ReceiveMessage(ctx context.Context) ([]port.Message, error) {
	partitionConsumer, err := k.Consumer.ConsumePartition(k.Config.Topic, 0, sarama.OffsetOldest)
	if err != nil {
		return nil, errors.Wrap(err, "error al crear la particion del consumidor")
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			k.Log.Error("Error al cerrar la particion del consumidor", zap.Error(err))
		}
	}()

	var messages []port.Message

	select {
	case msg := <-partitionConsumer.Messages():
		k.Log.Info("Mensaje recibido desde Kafka", zap.String("MessageID", string(msg.Key)))
		var message port.Message
		if err := json.Unmarshal(msg.Value, &message); err != nil {
			return nil, errors.Wrap(err, "error al deserializar mensaje")
		}
		messages = append(messages, message)
	case <-ctx.Done():
		return messages, ctx.Err()
	}
	return messages, nil
}

func (k *KafkaService) DeleteMessage(ctx context.Context, receiptHandle string) error {
	// Kafka no tiene el concepto de eliminar mensajes individuales
	// Por lo cual los mensajes se eliminan automaticamente segun el periodo de retencion

	k.Log.Info("Se llamo a DeleteMessage, pero kafka no admite la eliminacion de mensajes individuales",
		zap.String("receiptHandle", receiptHandle))
	return nil
}

func (k *KafkaService) Close() error {
	return k.Producer.Close()
}
