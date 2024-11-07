package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/port"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

const (
	maxRetries          = 3
	retryBaseDelay      = time.Second
	maxNumberOfMessages = 10
	waitTimeSeconds     = 20
)

type SQSService struct {
	Client port.SQSClientInterface
	Config *config.Config
	Log    logger.Logger
}

func NewSQSService(cfgApplication *config.Config, log logger.Logger) (*SQSService, error) {
	if cfgApplication == nil {
		return nil, errors.New("config no puede ser nil")
	}

	if log == nil {
		return nil, errors.New("logger no puede ser nil")
	}

	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(),
		awsCfg.WithRegion(cfgApplication.AWS.Region),
		awsCfg.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfgApplication.AWS.Credentials.AccessKey,
				cfgApplication.AWS.Credentials.SecretKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)

	return &SQSService{
		Client: sqsClient,
		Config: cfgApplication,
		Log:    log,
	}, nil
}

func (s *SQSService) SendMessage(ctx context.Context, message model.Message) error {
	body, err := json.Marshal(message)
	if err != nil {
		return errors.Wrap(err, "error al serializar mensaje")
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.Config.Messaging.SQS.QueueURL),
		MessageBody: aws.String(string(body)),
	}

	return s.sendMessageWithRetry(ctx, input, message.ID)
}

// sendMessageWithRetry implementa la lógica de reintento para enviar mensajes
func (s *SQSService) sendMessageWithRetry(ctx context.Context, input *sqs.SendMessageInput, messageID string) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := s.Client.SendMessage(ctx, input)
		if err == nil {
			s.Log.Info("Message enviado con exito",
				zap.String("messageID", messageID),
				zap.String("sqsMessageID", *result.MessageId))
			return nil
		}

		lastErr = err
		backoff := time.Duration(attempt+1) * retryBaseDelay

		s.Log.Warn("No se pudo enviar el mensaje, reintentando",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
			zap.Duration("backoff", backoff))

		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "contexto cancelado al reintentar")
		case <-time.After(backoff):
			continue
		}
	}

	return errors.Wrap(lastErr, "No se pudo enviar el mensaje a SQS después de todos los reintentos.")
}

func (s *SQSService) ReceiveMessage(ctx context.Context) ([]model.Message, error) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(s.Config.Messaging.SQS.QueueURL),
		MaxNumberOfMessages:   maxNumberOfMessages,
		WaitTimeSeconds:       waitTimeSeconds,
		AttributeNames:        []types.QueueAttributeName{types.QueueAttributeNameAll},
		MessageAttributeNames: []string{"All"},
	}

	var messages []model.Message
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := s.Client.ReceiveMessage(ctx, input)
		if err == nil {
			messages, err = s.processReceivedMessages(result.Messages)
			if err == nil {
				s.Log.Info("Message recibido con exito",
					zap.Int("count", len(result.Messages)))
				return messages, nil
			}
		}

		lastErr = err
		backoff := time.Duration(attempt+1) * retryBaseDelay

		s.Log.Warn("Error al recibir los mensajes, reintentando",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
			zap.Duration("backoff", backoff))

		select {
		case <-ctx.Done():
			return nil, errors.Wrap(ctx.Err(), "contexto cancelado al recibir mensajes")
		case <-time.After(backoff):
			continue
		}
	}
	return nil, errors.Wrap(lastErr, "No se pudieron recibir mensajes de SQS después de todos los reintentos.")
}

func (s *SQSService) processReceivedMessages(sqsMessages []types.Message) ([]model.Message, error) {
	messages := make([]model.Message, 0, len(sqsMessages))

	for _, sqsMsg := range sqsMessages {
		var msg model.Message
		err := json.Unmarshal([]byte(*sqsMsg.Body), &msg)
		if err != nil {
			s.Log.Error("No se pudo deserializar el mensaje",
				zap.Error(err),
				zap.String("messageID", *sqsMsg.MessageId))
			continue
		}

		msg.ReceiptHandle = *sqsMsg.ReceiptHandle
		messages = append(messages, msg)
	}

	return messages, nil
}

func (s *SQSService) DeleteMessage(ctx context.Context, receiptHandle string) error {
	if receiptHandle == "" {
		return errors.New("receipt handle no puede estar vacio")
	}

	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.Config.Messaging.SQS.QueueURL),
		ReceiptHandle: aws.String(receiptHandle),
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		_, err := s.Client.DeleteMessage(ctx, input)
		if err == nil {
			s.Log.Info("Mensaje eliminado con exito",
				zap.String("receiptHandle", receiptHandle))
			return nil
		}
		lastErr = err
		backoff := time.Duration(attempt+1) * retryBaseDelay

		s.Log.Warn("Error al eliminar el mensaje, reintentando",
			zap.Error(err),
			zap.String("receiptHandle", receiptHandle),
			zap.Int("attempt", attempt+1),
			zap.Duration("backoff", backoff))

		select {
		case <-ctx.Done():
			return errors.Wrap(ctx.Err(), "contexto cancelado mientras eliminaba el mensaje")
		case <-time.After(backoff):
			continue
		}
	}
	return errors.Wrap(lastErr, "No se pudo eliminar el mensaje de SQS después de todos los reintentos.")
}
