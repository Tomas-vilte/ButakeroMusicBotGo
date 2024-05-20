package message_queue

import (
	"context"
	"encoding/json"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.uber.org/zap"
)

// Publisher define un método para publicar un evento en una cola de mensajes.
type Publisher interface {
	Publish(ctx context.Context, event interface{}) error
}

// SQSPublisher es una implementación de Publisher que utiliza Amazon SQS para publicar eventos en una cola de mensajes.
type SQSPublisher struct {
	Client   SQSClient      // Client es el cliente SQS utilizado para enviar mensajes a la cola de mensajes.
	QueueURL string         // QueueURL es la URL de la cola de mensajes en la que se publicarán los eventos.
	Logger   logging.Logger // Logger se utiliza para registrar información y errores durante la publicación de eventos.
}

// NewSQSPublisher crea una nueva instancia de SQSPublisher con el cliente SQS, la URL de la cola y el logger dados.
func NewSQSPublisher(client SQSClient, queueURL string, logger logging.Logger) *SQSPublisher {
	return &SQSPublisher{
		Client:   client,
		QueueURL: queueURL,
		Logger:   logger,
	}
}

// Publish publica un evento en la cola de mensajes utilizando el cliente SQS y la URL de la cola proporcionados.
func (p *SQSPublisher) Publish(ctx context.Context, event interface{}) error {
	message, err := json.Marshal(event)
	if err != nil {
		p.Logger.Error("Error serializando el evento", zap.Error(err))
		return err
	}

	_, err = p.Client.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.QueueURL),
		MessageBody: aws.String(string(message)),
	})
	if err != nil {
		p.Logger.Error("Error enviando el mensaje a SQS", zap.Error(err))
		return err
	}
	p.Logger.Info("Mensaje enviado a SQS con exito", zap.String("QueueURL", p.QueueURL))
	return nil
}
