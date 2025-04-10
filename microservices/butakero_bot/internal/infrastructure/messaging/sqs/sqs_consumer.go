package sqs

import (
	"context"
	"encoding/json"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.uber.org/zap"
	"sync"
	"time"
)

type ClientSQS interface {
	ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error)
	DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error)
}

type SQSConsumer struct {
	client      ClientSQS
	logger      logging.Logger
	messageChan chan *entity.MessageQueue
	wg          sync.WaitGroup
	cfg         *config.Config
}

func NewSQSConsumer(client ClientSQS, config *config.Config, logger logging.Logger) *SQSConsumer {
	logger = logger.With(
		zap.String("component", "sqs_consumer"),
		zap.String("queueURL", config.QueueConfig.SQSConfig.Queues.BotDownloadStatusQueueURL),
		zap.Int32("maxMessages", config.QueueConfig.SQSConfig.MaxMessages),
		zap.Int32("waitTimeSeconds", config.QueueConfig.SQSConfig.WaitTimeSeconds),
	)

	return &SQSConsumer{
		client:      client,
		logger:      logger,
		cfg:         config,
		messageChan: make(chan *entity.MessageQueue),
	}
}

func (s *SQSConsumer) ConsumeMessages(ctx context.Context, _ int64) error {
	logger := s.logger.With(zap.String("method", "ConsumeMessages"))
	logger.Info("Iniciando consumo de mensajes SQS")

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer close(s.messageChan)

		for {
			select {
			case <-ctx.Done():
				logger.Info("Contexto cancelado, deteniendo consumidor SQS")
				return
			default:
				s.receiveAndProcessMessages(ctx)
			}
		}
	}()
	return nil
}

func (s *SQSConsumer) receiveAndProcessMessages(ctx context.Context) {
	logger := s.logger.With(zap.String("method", "receiveAndProcessMessages"))

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.cfg.QueueConfig.SQSConfig.Queues.BotDownloadStatusQueueURL),
		MaxNumberOfMessages: s.cfg.QueueConfig.SQSConfig.MaxMessages,
		WaitTimeSeconds:     s.cfg.QueueConfig.SQSConfig.WaitTimeSeconds,
		MessageSystemAttributeNames: []types.MessageSystemAttributeName{
			types.MessageSystemAttributeNameAll,
		},
		MessageAttributeNames: []string{"All"},
	}

	resp, err := s.client.ReceiveMessage(ctx, input)
	if err != nil {
		logger.Error("Error al recibir mensajes de la cola SQS", zap.Error(err))
		time.Sleep(time.Second)
		return
	}

	for _, msg := range resp.Messages {
		s.handleMessage(ctx, msg)
	}
}

func (s *SQSConsumer) handleMessage(ctx context.Context, msg types.Message) {
	logger := s.logger.With(
		zap.String("method", "handleMessage"),
		zap.String("messageID", *msg.MessageId),
	)

	logger.Debug("Mensaje recibido")

	var statusMessage entity.MessageQueue
	if err := json.Unmarshal([]byte(*msg.Body), &statusMessage); err != nil {
		logger.Error("Error al deserializar mensaje",
			zap.Error(err),
			zap.String("messageBody", *msg.Body))
		return
	}

	if statusMessage.Status == "success" {
		logger.Info("Mensaje procesado exitosamente",
			zap.String("status", statusMessage.Status),
			zap.String("messageId", *msg.MessageId))
		s.messageChan <- &statusMessage
	} else {
		logger.Warn("Mensaje recibido con estado de error",
			zap.Any("status", statusMessage),
			zap.String("messageId", *msg.MessageId))
	}

	s.deleteMessage(ctx, msg)
}

func (s *SQSConsumer) GetMessagesChannel() <-chan *entity.MessageQueue {
	return s.messageChan
}

func (s *SQSConsumer) deleteMessage(ctx context.Context, msg types.Message) {
	logger := s.logger.With(
		zap.String("method", "deleteMessage"),
		zap.String("messageID", *msg.MessageId),
	)

	deleteInput := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.cfg.QueueConfig.SQSConfig.Queues.BotDownloadStatusQueueURL),
		ReceiptHandle: msg.ReceiptHandle,
	}

	_, err := s.client.DeleteMessage(ctx, deleteInput)
	if err != nil {
		logger.Error("Error al eliminar mensaje de SQS", zap.Error(err))
	}
}
