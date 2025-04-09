package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"go.uber.org/zap"
)

type ProducerSQS struct {
	client *sqs.Client
	logger logger.Logger
	cfg    *config.Config
}

func NewProducerSQS(cfgApplication *config.Config, log logger.Logger) (ports.MessageProducer, error) {
	log.Info("Inicializando productor SQS",
		zap.String("region", cfgApplication.AWS.Region),
		zap.String("queue_url", cfgApplication.Messaging.SQS.QueueURL))

	if cfgApplication == nil {
		log.Error("Error: configuración inválida", zap.Error(errors.ErrInvalidInput))
		return nil, errors.ErrInvalidInput.WithMessage("config no puede ser nil")
	}

	if log == nil {
		return nil, errors.ErrInvalidInput.WithMessage("logger no puede ser nil")
	}

	log.Debug("Cargando configuración AWS", zap.String("region", cfgApplication.AWS.Region))
	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(),
		awsCfg.WithRegion(cfgApplication.AWS.Region))
	if err != nil {
		log.Error("Error cargando configuración AWS", zap.Error(err))
		return nil, errors.ErrCodeDBConnectionFailed.WithMessage(fmt.Sprintf("error cargando configuración AWS: %v", err))
	}

	log.Debug("Creando cliente SQS")
	sqsClient := sqs.NewFromConfig(cfg)

	log.Info("Productor SQS inicializado correctamente",
		zap.String("queue_url", cfgApplication.Messaging.SQS.QueueURL))
	return &ProducerSQS{
		client: sqsClient,
		cfg:    cfgApplication,
		logger: log,
	}, nil
}

func (p *ProducerSQS) Publish(ctx context.Context, msg *model.MediaProcessingMessage) error {
	log := p.logger.With(
		zap.String("component", "sqs_producer"),
		zap.String("method", "Publish"),
		zap.String("video_id", msg.VideoID),
		zap.String("status", msg.Status),
	)

	log.Debug("Serializando mensaje para publicar")
	body, err := json.Marshal(msg)
	if err != nil {
		log.Error("Error serializando mensaje", zap.Error(err))
		return err
	}

	log.Debug("Mensaje serializado", zap.Int("payload_size", len(body)))

	input := &sqs.SendMessageInput{
		MessageBody: aws.String(string(body)),
		QueueUrl:    aws.String(p.cfg.Messaging.SQS.QueueURL),
	}

	log.Debug("Enviando mensaje a SQS",
		zap.String("queue_url", p.cfg.Messaging.SQS.QueueURL))
	result, err := p.client.SendMessage(ctx, input)
	if err != nil {
		log.Error("Error publicando mensaje en SQS",
			zap.String("queue_url", p.cfg.Messaging.SQS.QueueURL),
			zap.Error(err))
		return err
	}

	log.Info("Mensaje publicado exitosamente en SQS",
		zap.String("message_id", *result.MessageId),
		zap.String("video_id", msg.VideoID),
		zap.String("status", msg.Status),
		zap.Bool("success", msg.Success))

	if result.SequenceNumber != nil {
		log.Debug("Detalles adicionales del mensaje",
			zap.String("sequence_number", *result.SequenceNumber))
	}

	return nil
}

func (p *ProducerSQS) Close() error {
	log := p.logger.With(
		zap.String("component", "sqs_producer"),
		zap.String("method", "Close"),
	)

	log.Info("Cerrando productor SQS")
	log.Info("Productor SQS cerrado correctamente")
	return nil
}
