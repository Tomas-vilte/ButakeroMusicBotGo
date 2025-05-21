package sqs

import (
	"context"
	"encoding/json"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/zap"
	"time"
)

type ConsumerSQS struct {
	cfg    *config.Config
	logger logger.Logger
	client *sqs.Client
}

func NewConsumerSQS(cfgApplication *config.Config, log logger.Logger) (ports.MessageConsumer, error) {
	log.Info("Inicializando consumidor SQS",
		zap.String("queue_url", cfgApplication.Messaging.SQS.QueueURLs.BotDownloadRequestsURL))

	log.Debug("Cargando configuración AWS", zap.String("region", cfgApplication.AWS.Region))
	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(),
		awsCfg.WithRegion(cfgApplication.AWS.Region))
	if err != nil {
		log.Error("Error cargando configuración AWS", zap.Error(err))
		return nil, errors.ErrSQSAWSConfig.Wrap(err)
	}

	log.Debug("Creando cliente SQS")
	sqsClient := sqs.NewFromConfig(cfg)

	log.Info("Consumidor SQS inicializado correctamente",
		zap.String("queue_url", cfgApplication.Messaging.SQS.QueueURLs.BotDownloadRequestsURL))
	return &ConsumerSQS{
		client: sqsClient,
		cfg:    cfgApplication,
		logger: log,
	}, nil
}

func (c *ConsumerSQS) GetRequestsChannel(ctx context.Context) (<-chan *model.MediaRequest, error) {
	log := c.logger.With(
		zap.String("component", "sqs_consumer"),
		zap.String("method", "GetRequestsChannel"),
		zap.String("queue_url", c.cfg.Messaging.SQS.QueueURLs.BotDownloadRequestsURL),
	)

	log.Info("Iniciando canal de solicitudes SQS")
	out := make(chan *model.MediaRequest)

	log.Debug("Iniciando rutina de consumo SQS")
	go c.consumeLoop(ctx, out)

	log.Info("Canal de solicitudes SQS creado correctamente")
	return out, nil
}

func (c *ConsumerSQS) consumeLoop(ctx context.Context, out chan<- *model.MediaRequest) {
	log := c.logger.With(
		zap.String("component", "sqs_consumer"),
		zap.String("method", "consumeLoop"),
		zap.String("queue_url", c.cfg.Messaging.SQS.QueueURLs.BotDownloadRequestsURL),
	)

	log.Info("Loop de consumo SQS iniciado")
	defer func() {
		log.Debug("Cerrando canal de salida")
		close(out)
		log.Info("Loop de consumo SQS finalizado")
	}()

	pollCount := 0
	for {
		select {
		case <-ctx.Done():
			log.Info("Contexto cancelado, deteniendo consumidor SQS")
			return
		default:
			pollCount++
			log.Debug("Realizando polling de mensajes", zap.Int("poll_count", pollCount))
			c.receiveMessages(ctx, out)

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (c *ConsumerSQS) receiveMessages(ctx context.Context, out chan<- *model.MediaRequest) {
	log := c.logger.With(
		zap.String("component", "sqs_consumer"),
		zap.String("method", "receiveMessages"),
		zap.String("queue_url", c.cfg.Messaging.SQS.QueueURLs.BotDownloadRequestsURL),
	)

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.cfg.Messaging.SQS.QueueURLs.BotDownloadRequestsURL),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     20,
		MessageSystemAttributeNames: []types.MessageSystemAttributeName{
			"All",
		},
		MessageAttributeNames: []string{"All"},
	}

	log.Debug("Solicitando mensajes de SQS",
		zap.Int32("max_messages", input.MaxNumberOfMessages),
		zap.Int32("wait_time_seconds", input.WaitTimeSeconds))

	result, err := c.client.ReceiveMessage(ctx, input)
	if err != nil {
		log.Error("Error recibiendo mensajes de SQS", zap.Error(errors.ErrSQSMessageConsume.Wrap(err)))
		return
	}

	messageCount := len(result.Messages)
	if messageCount > 0 {
		log.Info("Mensajes recibidos de SQS", zap.Int("count", messageCount))
	} else {
		log.Debug("No se recibieron mensajes de SQS")
		return
	}

	for i, msg := range result.Messages {
		log.Debug("Procesando mensaje",
			zap.Int("index", i),
			zap.String("message_id", *msg.MessageId))

		var req model.MediaRequest
		if err := json.Unmarshal([]byte(*msg.Body), &req); err != nil {
			log.Error("Error deserializando mensaje SQS",
				zap.String("message_id", *msg.MessageId),
				zap.Error(errors.ErrSQSMessageDeserialize.Wrap(err)),
				zap.String("body", *msg.Body))
			continue
		}

		log.Info("Mensaje SQS procesado correctamente",
			zap.String("message_id", *msg.MessageId),
			zap.String("request_id", req.RequestID),
			zap.String("user_id", req.UserID),
			zap.String("provider_type", req.ProviderType))

		if len(msg.MessageAttributes) > 0 {
			log.Debug("Atributos del mensaje",
				zap.Int("attribute_count", len(msg.MessageAttributes)))
		}

		out <- &req

		log.Debug("Eliminando mensaje procesado",
			zap.String("message_id", *msg.MessageId),
			zap.String("receipt_handle", *msg.ReceiptHandle))
		c.deleteMessage(ctx, msg.ReceiptHandle)
	}
}

func (c *ConsumerSQS) deleteMessage(ctx context.Context, receiptHandle *string) {
	log := c.logger.With(
		zap.String("component", "sqs_consumer"),
		zap.String("method", "deleteMessage"),
		zap.String("receipt_handle", *receiptHandle),
	)

	log.Debug("Enviando solicitud de eliminación de mensaje")
	_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.cfg.Messaging.SQS.QueueURLs.BotDownloadRequestsURL),
		ReceiptHandle: receiptHandle,
	})

	if err != nil {
		log.Error("Error eliminando mensaje de SQS",
			zap.String("receipt_handle", *receiptHandle),
			zap.Error(errors.ErrSQSMessageDelete.Wrap(err)))
		return
	}

	log.Debug("Mensaje eliminado correctamente")
}

func (c *ConsumerSQS) Close() error {
	log := c.logger.With(
		zap.String("component", "sqs_consumer"),
		zap.String("method", "Close"),
	)

	log.Info("Cerrando consumidor SQS")
	log.Info("Consumidor SQS cerrado correctamente")
	return nil
}
