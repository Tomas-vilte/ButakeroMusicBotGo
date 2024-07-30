package sqs

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.uber.org/zap"
)

type (
	sqsClient struct {
		sqsClient *sqs.SQS
		queueURL  string
		logger    logging.Logger
	}

	SQSClient interface {
		SendMessage(ctx context.Context, messageBody string) error
	}
)

func NewSQSClient(sess *session.Session, queueURL string, logger logging.Logger) SQSClient {
	return &sqsClient{
		sqsClient: sqs.New(sess),
		queueURL:  queueURL,
		logger:    logger,
	}
}

func (c *sqsClient) SendMessage(ctx context.Context, messageBody string) error {
	_, err := c.sqsClient.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		MessageBody: aws.String(messageBody),
		QueueUrl:    aws.String(c.queueURL),
	})
	if err != nil {
		c.logger.Error("Error al enviar el mensaje a SQS", zap.Error(err))
		return err
	}
	c.logger.Info("Mensaje enviado con exito a SQS")
	return nil
}
