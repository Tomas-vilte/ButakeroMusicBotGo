package github_event

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/common"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/logging"
	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
	"net/http"
)

const (
	// ReleaseAction representa la acción de publicación de un lanzamiento en GitHub.
	ReleaseAction = "published"
	// WorkflowJobAction representa la acción de finalización de un trabajo de flujo de trabajo en GitHub.
	WorkflowJobAction = "completed"
)

// EventProcessor define un método para procesar un evento común.
type EventProcessor interface {
	ProcessEvent(ctx context.Context, event interface{}) error
}

// JSONMarshaller define un método para serializar datos a JSON.
type JSONMarshaller interface {
	Marshal(v interface{}) ([]byte, error)
}

// GitHubEventDecoder define un método para decodificar eventos recibidos de GitHub.
type GitHubEventDecoder interface {
	DecodeGitHubEvent(body string) (interface{}, error)
}

// EventHandler maneja eventos recibidos desde GitHub.
type EventHandler struct {
	Processor  EventProcessor // Processor es el procesador de eventos utilizado para procesar los eventos recibidos.
	Logger     logging.Logger // Logger registra información y errores durante el manejo de eventos.
	Decoder    GitHubEventDecoder
	Marshaller JSONMarshaller
}

// NewEventHandler crea una nueva instancia de EventHandler con el procesador de eventos y el logger dados.
func NewEventHandler(processor EventProcessor, logger logging.Logger, decoder GitHubEventDecoder, marshaller JSONMarshaller) *EventHandler {
	return &EventHandler{
		Processor:  processor,
		Logger:     logger,
		Decoder:    decoder,
		Marshaller: marshaller,
	}
}

// HandleGitHubEvent maneja un evento recibido desde GitHub.
// Decodifica el evento recibido, lo procesa y devuelve una respuesta adecuada.
func (h *EventHandler) HandleGitHubEvent(ctx context.Context, requests events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	event, err := h.Decoder.DecodeGitHubEvent(requests.Body)
	if err != nil {
		h.Logger.Error("Error al decodificar el evento", zap.Error(err))
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest, Body: "Error al decodificar el evento"}, err
	}

	err = h.Processor.ProcessEvent(ctx, event)
	if err != nil {
		h.Logger.Error("Error al procesar el evento", zap.Error(err))
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError, Body: "Error al procesar el evento: " + err.Error()}, err
	}

	responseJSON, err := h.Marshaller.Marshal(event)
	if err != nil {
		h.Logger.Error("Error al codificar la respuesta a JSON", zap.Error(err))
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError, Body: "Error al codificar la respuesta a JSON"}, err
	}

	return events.APIGatewayProxyResponse{StatusCode: http.StatusOK, Body: string(responseJSON)}, nil
}

// DefaultGitHubEventDecoder implementa GitHubEventDecoder utilizando una decodificación predeterminada.
type DefaultGitHubEventDecoder struct {
	Logger logging.Logger
}

// NewGitHubEventDecoder crea una nueva instancia de DefaultGitHubEventDecoder con el logger dado.
func NewGitHubEventDecoder(logger logging.Logger) *DefaultGitHubEventDecoder {
	return &DefaultGitHubEventDecoder{
		Logger: logger,
	}
}

// DecodeGitHubEvent decodifica un evento recibido de GitHub y lo devuelve como una interfaz genérica.
func (h *DefaultGitHubEventDecoder) DecodeGitHubEvent(body string) (interface{}, error) {
	var event interface{}
	if err := json.Unmarshal([]byte(body), &event); err != nil {
		h.Logger.Error("Error al decodificar el evento", zap.Error(err))
		return common.Event{}, err
	}
	switch event.(map[string]interface{})["action"] {
	case ReleaseAction:
		var releaseEvent common.ReleaseEvent
		if err := json.Unmarshal([]byte(body), &releaseEvent); err != nil {
			return nil, err
		}
		return releaseEvent, nil
	case WorkflowJobAction:
		var workflowEvent common.WorkflowEvent
		if err := json.Unmarshal([]byte(body), &workflowEvent); err != nil {
			return nil, err
		}
		return workflowEvent, nil
	default:
		h.Logger.Error("Acción de evento no reconocida")
		return nil, errors.New("acción de evento no reconocida")
	}
}

// DefaultJSONMarshaller implementa JSONMarshaller utilizando la codificación predeterminada de JSON.
type DefaultJSONMarshaller struct {
}

// Marshal serializa los datos a JSON utilizando la codificación predeterminada de JSON.
func (m *DefaultJSONMarshaller) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// NewDefaultJSONMarshaller crea una nueva instancia de DefaultJSONMarshaller.
func NewDefaultJSONMarshaller() *DefaultJSONMarshaller {
	return &DefaultJSONMarshaller{}
}
