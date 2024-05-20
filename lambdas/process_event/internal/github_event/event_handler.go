package github_event

import (
	"context"
	"encoding/json"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/common"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/logging"
	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
	"net/http"
)

// EventProcessor define un método para procesar un evento común.
type EventProcessor interface {
	ProcessEvent(ctx context.Context, event common.Event) error
}

// EventHandler maneja eventos recibidos desde GitHub.
type EventHandler struct {
	Processor EventProcessor // Processor es el procesador de eventos utilizado para procesar los eventos recibidos.
	Logger    logging.Logger // Logger registra información y errores durante el manejo de eventos.
}

// NewEventHandler crea una nueva instancia de EventHandler con el procesador de eventos y el logger dados.
func NewEventHandler(processor EventProcessor, logger logging.Logger) *EventHandler {
	return &EventHandler{
		Processor: processor,
		Logger:    logger,
	}
}

// HandleGitHubEvent maneja un evento recibido desde GitHub.
// Decodifica el evento recibido, lo procesa y devuelve una respuesta adecuada.
func (h *EventHandler) HandleGitHubEvent(ctx context.Context, requests events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var event common.Event
	if err := json.Unmarshal([]byte(requests.Body), &event); err != nil {
		h.Logger.Error("Error al decodificar el evento", zap.Error(err))
		return events.APIGatewayProxyResponse{StatusCode: http.StatusOK}, err
	}

	h.Logger.Info("Evento recibido de Github", zap.String("Action", event.Action), zap.String("Tag", event.Release.TagName))
	err := h.Processor.ProcessEvent(ctx, event)
	if err != nil {
		h.Logger.Error("Error al procesar evento", zap.Error(err))
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError, Body: "Hubo un error en procesar el evento: " + err.Error()}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK, Body: "Evento Procesado con exito :D"}, nil
}
