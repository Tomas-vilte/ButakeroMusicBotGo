package service

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/message_queue"
	"go.uber.org/zap"
)

// EventProcessor representa un procesador de eventos encargado de procesar y publicar eventos.
type EventProcessor struct {
	SQSPublisher message_queue.Publisher // SQSPublisher es el encargado de publicar eventos en la cola de mensajes.
	Logger       logging.Logger          // Logger registra información y errores durante el procesamiento de eventos.
}

// NewEventProcessor crea una nueva instancia de EventProcessor con los parámetros indicados.
func NewEventProcessor(publisher message_queue.Publisher, logger logging.Logger) *EventProcessor {
	return &EventProcessor{
		SQSPublisher: publisher,
		Logger:       logger,
	}
}

// ProcessEvent procesa un evento dado y lo publica en la cola de mensajes.
// Utiliza el contexto proporcionado y devuelve un error si ocurre algún problema durante el procesamiento.
func (p *EventProcessor) ProcessEvent(ctx context.Context, event interface{}) error {

	err := p.SQSPublisher.Publish(ctx, event)
	if err != nil {
		p.Logger.Error("Error publicando el evento a SQS", zap.Error(err))
		return err
	}
	p.Logger.Info("Evento publicado en SQS con exito")
	return nil
}
