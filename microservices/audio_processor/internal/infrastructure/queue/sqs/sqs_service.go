package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
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
	Client *sqs.Client
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
		awsCfg.WithRegion(cfgApplication.AWS.Region))
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

func (s *SQSService) SendMessage(ctx context.Context, message *model.MediaProcessingMessage) error {
	log := s.Log.With(
		zap.String("component", "SQSService"),
		zap.String("method", "SendMessage"),
		zap.String("message_id", message.ID),
	)
	body, err := json.Marshal(message)
	if err != nil {
		log.Error("Error al serializar el mensaje", zap.Error(err))
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
	log := s.Log.With(
		zap.String("component", "SQSService"),
		zap.String("method", "sendMessageWithRetry"),
		zap.String("message_id", messageID),
	)
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := s.Client.SendMessage(ctx, input)
		if err == nil {
			log.Info("Mensaje enviado con éxito",
				zap.String("sqs_message_id", *result.MessageId))
			return nil
		}

		lastErr = err
		backoff := time.Duration(attempt+1) * retryBaseDelay

		log.Warn("Error al enviar el mensaje, reintentando",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
			zap.Duration("backoff", backoff))

		select {
		case <-ctx.Done():
			log.Error("Contexto cancelado durante el reintento", zap.Error(ctx.Err()))
			return errors.Wrap(ctx.Err(), "contexto cancelado al reintentar")
		case <-time.After(backoff):
			continue
		}
	}

	log.Error("No se pudo enviar el mensaje después de todos los reintentos", zap.Error(lastErr))
	return errors.Wrap(lastErr, "No se pudo enviar el mensaje a SQS después de todos los reintentos.")
}

func (s *SQSService) ReceiveMessage(ctx context.Context) ([]model.MediaProcessingMessage, error) {
	log := s.Log.With(
		zap.String("component", "SQSService"),
		zap.String("method", "ReceiveMessage"),
	)
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.Config.Messaging.SQS.QueueURL),
		MaxNumberOfMessages: maxNumberOfMessages,
		WaitTimeSeconds:     waitTimeSeconds,
		MessageSystemAttributeNames: []types.MessageSystemAttributeName{
			"All",
		},
		MessageAttributeNames: []string{"All"},
	}

	var messages []model.MediaProcessingMessage
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		result, err := s.Client.ReceiveMessage(ctx, input)
		if err == nil {
			messages, err = s.processReceivedMessages(result.Messages)
			if err == nil {
				log.Info("Mensajes recibidos con éxito",
					zap.Int("count", len(result.Messages)))
				return messages, nil
			}
		}

		lastErr = err
		backoff := time.Duration(attempt+1) * retryBaseDelay

		log.Warn("Error al recibir mensajes, reintentando",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
			zap.Duration("backoff", backoff))

		select {
		case <-ctx.Done():
			log.Error("Contexto cancelado durante la recepción de mensajes", zap.Error(ctx.Err()))
			return nil, errors.Wrap(ctx.Err(), "contexto cancelado al recibir mensajes")
		case <-time.After(backoff):
			continue
		}
	}

	log.Error("No se pudieron recibir mensajes después de todos los reintentos", zap.Error(lastErr))
	return nil, errors.Wrap(lastErr, "No se pudieron recibir mensajes de SQS después de todos los reintentos.")
}

func (s *SQSService) processReceivedMessages(sqsMessages []types.Message) ([]model.MediaProcessingMessage, error) {
	log := s.Log.With(
		zap.String("component", "SQSService"),
		zap.String("method", "processReceivedMessages"),
	)
	messages := make([]model.MediaProcessingMessage, 0, len(sqsMessages))

	for _, sqsMsg := range sqsMessages {
		var msg model.MediaProcessingMessage
		err := json.Unmarshal([]byte(*sqsMsg.Body), &msg)
		if err != nil {
			log.Error("Error al deserializar el mensaje",
				zap.Error(err),
				zap.String("sqs_message_id", *sqsMsg.MessageId))
			continue
		}

		msg.ReceiptHandle = *sqsMsg.ReceiptHandle
		messages = append(messages, msg)
	}

	log.Info("Mensajes procesados exitosamente", zap.Int("count", len(messages)))
	return messages, nil
}

func (s *SQSService) DeleteMessage(ctx context.Context, receiptHandle string) error {
	log := s.Log.With(
		zap.String("component", "SQSService"),
		zap.String("method", "DeleteMessage"),
		zap.String("receipt_handle", receiptHandle),
	)
	if receiptHandle == "" {
		log.Error("Receipt handle no puede estar vacío")
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
			log.Info("Mensaje eliminado con éxito")
			return nil
		}
		lastErr = err
		backoff := time.Duration(attempt+1) * retryBaseDelay

		log.Warn("Error al eliminar el mensaje, reintentando",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
			zap.Duration("backoff", backoff))

		select {
		case <-ctx.Done():
			log.Error("Contexto cancelado durante la eliminación del mensaje", zap.Error(ctx.Err()))
			return errors.Wrap(ctx.Err(), "contexto cancelado mientras eliminaba el mensaje")
		case <-time.After(backoff):
			continue
		}
	}

	log.Error("No se pudo eliminar el mensaje después de todos los reintentos", zap.Error(lastErr))
	return errors.Wrap(lastErr, "No se pudo eliminar el mensaje de SQS después de todos los reintentos.")
}
