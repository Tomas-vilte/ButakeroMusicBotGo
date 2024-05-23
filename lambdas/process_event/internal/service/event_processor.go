package service

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/common"
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
func (p *EventProcessor) ProcessEvent(ctx context.Context, event interface{}, eventType string) error {

	if err := p.validateEvent(event); err != nil {
		p.Logger.Error("Error de validación del evento", zap.Error(err))
		return err
	}

	err := p.SQSPublisher.Publish(ctx, event, eventType)
	if err != nil {
		p.Logger.Error("Error publicando el evento a SQS", zap.Error(err))
		return err
	}
	p.Logger.Info("Evento publicado en SQS con exito")
	return nil
}

func (p *EventProcessor) validateEvent(event interface{}) error {

	switch evt := event.(type) {
	case common.ReleaseEvent:
		if evt.Release.Name == "" {
			return errors.New("el campo 'Name' en el evento de lanzamiento está vacío")
		}
	case common.WorkflowEvent:
		if evt.WorkFlowJobs.WorkFlowName == "" {
			return errors.New("el campo 'Name' en el evento de flujo de trabajo está vacío")
		}
	default:
		return errors.New("tipo de evento no compatible para validación")
	}
	return nil
}
