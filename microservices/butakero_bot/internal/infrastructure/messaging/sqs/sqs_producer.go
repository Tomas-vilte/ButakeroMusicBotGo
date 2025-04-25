package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model/queue"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
)

type ProducerSQS struct {
	client *sqs.Client
	cfg    *config.Config
	logger logging.Logger
}

func NewProducerSQS(cfgApplication *config.Config, log logging.Logger) (ports.SongDownloadRequestPublisher, error) {
	log.Info("Inicializando productor SQS",
		zap.String("queue_url", cfgApplication.QueueConfig.SQSConfig.Queues.BotDownloadRequestsQueueURL))

	log.Debug("Cargando configuración AWS", zap.String("region", cfgApplication.AWS.Region))
	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(),
		awsCfg.WithRegion(cfgApplication.AWS.Region))
	if err != nil {
		log.Error("Error cargando configuración AWS", zap.Error(err))
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	log.Debug("Creando cliente SQS")
	sqsClient := sqs.NewFromConfig(cfg)

	log.Info("Productor SQS inicializado correctamente",
		zap.String("queue_url", cfgApplication.QueueConfig.SQSConfig.Queues.BotDownloadRequestsQueueURL))
	return &ProducerSQS{
		client: sqsClient,
		cfg:    cfgApplication,
		logger: log,
	}, nil
}

func (p *ProducerSQS) PublishDownloadRequest(ctx context.Context, message *queue.DownloadRequestMessage) error {
	log := p.logger.With(
		zap.String("component", "ProducerSQS"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "PublishDownloadRequest"),
		zap.String("queue_url", p.cfg.QueueConfig.SQSConfig.Queues.BotDownloadRequestsQueueURL),
	)

	select {
	case <-ctx.Done():
		return fmt.Errorf("contexto cancelado antes de publicar: %w", ctx.Err())
	default:
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Error("Error al serializar el mensaje", zap.Error(err))
		return fmt.Errorf("error al serializar el mensaje: %w", err)
	}

	msg := &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.cfg.QueueConfig.SQSConfig.Queues.BotDownloadRequestsQueueURL),
		MessageBody: aws.String(string(jsonData)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"RequestID": {
				DataType:    aws.String("String"),
				StringValue: aws.String(message.RequestID),
			},
		},
	}

	done := make(chan error, 1)
	go func() {
		resp, err := p.client.SendMessage(ctx, msg)
		if err != nil {
			done <- err
			return
		}
		log.Info("Mensaje publicado exitosamente",
			zap.String("message_id", *resp.MessageId),
			zap.String("interaction_id", message.RequestID))
		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("operación cancelada mientras se publicaba: %w", ctx.Err())
	}
}

func (p *ProducerSQS) ClosePublisher() error {
	return nil
}
