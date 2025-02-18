package sqs

import (
	"context"
	"encoding/json"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
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
	queueURL    string
	logger      logging.Logger
	messageChan chan *entity.StatusMessage
	maxMessages int32
	waitTime    int32
	wg          sync.WaitGroup
}

func NewSQSConsumer(client ClientSQS, config SQSConfig, logger logging.Logger) *SQSConsumer {
	return &SQSConsumer{
		client:      client,
		queueURL:    config.QueueURL,
		logger:      logger,
		messageChan: make(chan *entity.StatusMessage),
		maxMessages: config.MaxMessages,
		waitTime:    config.WaitTimeSeconds,
	}
}

func (s *SQSConsumer) ConsumeMessages(ctx context.Context, offset int64) error {
	s.logger.Info("Iniciando consumo de mensajes SQS",
		zap.String("queueURL", s.queueURL),
		zap.Int32("maxMessages", s.maxMessages),
		zap.Int32("waitTimeSeconds", s.waitTime))

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer close(s.messageChan)

		for {
			select {
			case <-ctx.Done():
				s.logger.Info("Contexto cancelado, deteniendo consumidor SQS")
				return
			default:
				s.receiveAndProcessMessages(ctx)
			}
		}
	}()
	return nil
}

func (s *SQSConsumer) receiveAndProcessMessages(ctx context.Context) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.queueURL),
		MaxNumberOfMessages: s.maxMessages,
		WaitTimeSeconds:     s.waitTime,
		MessageSystemAttributeNames: []types.MessageSystemAttributeName{
			types.MessageSystemAttributeNameAll,
		},
		MessageAttributeNames: []string{"All"},
	}

	resp, err := s.client.ReceiveMessage(ctx, input)
	if err != nil {
		s.logger.Error("Error al recibir mensajes de la cola SQS", zap.Error(err))
		time.Sleep(time.Second)
		return
	}

	for _, msg := range resp.Messages {
		s.handleMessage(ctx, msg)
	}
}

func (s *SQSConsumer) handleMessage(ctx context.Context, msg types.Message) {
	s.logger.Debug("Mensaje recibido", zap.String("messageID", *msg.MessageId))

	var statusMessage entity.StatusMessage
	if err := json.Unmarshal([]byte(*msg.Body), &statusMessage); err != nil {
		s.logger.Error("Error al deserializar mensaje",
			zap.Error(err),
			zap.String("messageBody", *msg.Body))
		return
	}

	if statusMessage.Status.Status == "success" {
		s.logger.Info("Mensaje procesado exitosamente",
			zap.String("status", statusMessage.Status.Status),
			zap.String("messageId", *msg.MessageId))
		s.messageChan <- &statusMessage
	} else {
		s.logger.Warn("Mensaje recibido con estado de error",
			zap.Any("status", statusMessage),
			zap.String("messageId", *msg.MessageId))
	}

	s.deleteMessage(ctx, msg)
}

func (s *SQSConsumer) GetMessagesChannel() <-chan *entity.StatusMessage {
	return s.messageChan
}

func (s *SQSConsumer) deleteMessage(ctx context.Context, msg types.Message) {
	deleteInput := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.queueURL),
		ReceiptHandle: msg.ReceiptHandle,
	}

	_, err := s.client.DeleteMessage(ctx, deleteInput)
	if err != nil {
		s.logger.Error("Error al eliminar mensaje de SQS",
			zap.Error(err),
			zap.String("messageId", *msg.MessageId))
	}
}
