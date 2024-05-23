package message_queue

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.uber.org/zap"
)

// Publisher define un método para publicar un evento en una cola de mensajes.
type Publisher interface {
	Publish(ctx context.Context, event interface{}, eventType string) error
}

// SQSPublisher es una implementación de Publisher que utiliza Amazon SQS para publicar eventos en una cola de mensajes.
type SQSPublisher struct {
	Client          SQSClient         // Client es el cliente SQS utilizado para enviar mensajes a la cola de mensajes.
	QueueURLsByType map[string]string // QueueURLsByType es un mapa que asocia tipos de eventos con URLs de cola de mensajes.
	Logger          logging.Logger    // Logger se utiliza para registrar información y errores durante la publicación de eventos.
}

// NewSQSPublisher crea una nueva instancia de SQSPublisher con el cliente SQS, la URL de la cola y el logger dados.
func NewSQSPublisher(client SQSClient, queueURLsByType map[string]string, logger logging.Logger) *SQSPublisher {
	return &SQSPublisher{
		Client:          client,
		QueueURLsByType: queueURLsByType,
		Logger:          logger,
	}
}

// Publish publica un evento en la cola de mensajes utilizando el cliente SQS y la URL de la cola proporcionados.
func (p *SQSPublisher) Publish(ctx context.Context, event interface{}, eventType string) error {
	message, err := json.Marshal(event)
	if err != nil {
		p.Logger.Error("Error serializando el evento", zap.Error(err))
		return err
	}

	queueURL, ok := p.QueueURLsByType[eventType]
	if !ok {
		p.Logger.Error("No se encontró URL de cola para el tipo de evento", zap.String("eventType", eventType))
		return errors.New("URL de cola no encontrada para el tipo de evento: " + eventType)
	}

	_, err = p.Client.SendMessageWithContext(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(message)),
	})
	if err != nil {
		p.Logger.Error("Error enviando el mensaje a SQS", zap.Error(err))
		return err
	}
	p.Logger.Info("Mensaje enviado a SQS con exito", zap.String("QueueURL", queueURL))
	return nil
}
