package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

const (
	maxRetries     = 3
	retryBaseDelay = time.Second
)

type SQSService struct {
	Client queue.SQSClientInterface
	Config config.Config
	Log    logger.Logger
}

func NewSQSService(cfgApplication config.Config, log logger.Logger) (*SQSService, error) {
	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(), awsCfg.WithRegion(cfgApplication.Region), awsCfg.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(
			cfgApplication.AccessKey, cfgApplication.SecretKey, "")))
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

// SendMessage envía un mensaje a la cola SQS.
func (s *SQSService) SendMessage(ctx context.Context, message queue.Message) error {
	body, err := json.Marshal(message)
	if err != nil {
		return errors.Wrap(err, "error al serializar mensaje")
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.Config.QueueURL),
		MessageBody: aws.String(string(body)),
	}

	for i := 0; i < maxRetries; i++ {
		_, err = s.Client.SendMessage(ctx, input)
		if err == nil {
			s.Log.Info("Mensaje enviado exitosamente", zap.String("messageID", message.ID))
			return nil
		}
		s.Log.Warn("Error al enviar mensaje, reintentando", zap.Error(err), zap.Int("retry", i+1))
		time.Sleep(retryBaseDelay * time.Duration(i+1))
	}

	return errors.Wrap(err, "error al enviar mensaje a SQS después de varios intentos")
}

// ReceiveMessage recibe un mensaje de la cola SQS.
func (s *SQSService) ReceiveMessage(ctx context.Context) (*types.Message, error) {
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(s.Config.QueueURL),
		MaxNumberOfMessages: 1,
		WaitTimeSeconds:     20,
	}

	output, err := s.Client.ReceiveMessage(ctx, input)
	if err != nil {
		return nil, errors.Wrap(err, "error al recibir mensaje de SQS")
	}

	if len(output.Messages) == 0 {
		return nil, nil
	}

	return &output.Messages[0], nil
}

// DeleteMessage elimina un mensaje de la cola SQS.
func (s *SQSService) DeleteMessage(ctx context.Context, receiptHandle string) error {
	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(s.Config.QueueURL),
		ReceiptHandle: aws.String(receiptHandle),
	}

	_, err := s.Client.DeleteMessage(ctx, input)
	if err != nil {
		return errors.Wrap(err, "error al eliminar el mensaje de SQS")
	}
	return nil
}
