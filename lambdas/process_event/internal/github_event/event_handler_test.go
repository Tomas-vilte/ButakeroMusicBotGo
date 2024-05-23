package github_event

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/common"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"net/http"
	"testing"
)

type MockEventProcessor struct {
	mock.Mock
}

func (m *MockEventProcessor) ProcessEvent(ctx context.Context, event interface{}, eventType string) error {
	args := m.Called(ctx, event, eventType)
	return args.Error(0)
}

type MockLogger struct {
	mock.Mock
}

// Mock para GitHubEventDecoder
type MockGitHubEventDecoder struct {
	mock.Mock
}

func (m *MockGitHubEventDecoder) DecodeGitHubEvent(body string) (interface{}, error) {
	args := m.Called(body)
	return args.Get(0), args.Error(1)
}

func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

type MockJSONMarshaler struct {
	mock.Mock
}

func (m *MockJSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	args := m.Called(v)
	return args.Get(0).([]byte), args.Error(1)
}

func TestHandleGitHubEvent(t *testing.T) {
	// Configurar mocks
	eventProcessorMock := new(MockEventProcessor)
	decoderMock := new(MockGitHubEventDecoder)
	logger := new(MockLogger)
	jsonMarshallerMock := new(MockJSONMarshaler)

	// Configurar instancia de EventHandler con los mocks
	handler := NewEventHandler(eventProcessorMock, logger, decoderMock, jsonMarshallerMock)

	// Configurar datos de prueba
	request := events.APIGatewayProxyRequest{
		Body: "{\"action\": \"published\"}",
	}

	// Simular el comportamiento del decoderMock
	expectedEvent := common.ReleaseEvent{
		Action: "published",
		Release: common.Release{
			TagName: "v1.0.0",
			Name:    "Initial Release",
		},
	}
	decoderMock.On("DecodeGitHubEvent", request.Body).Return(expectedEvent, nil)

	// Simular el procesamiento del evento
	eventProcessorMock.On("ProcessEvent", mock.Anything, expectedEvent, "release").Return(nil)

	// Simular la codificación de la respuesta JSON
	expectedJSON := []byte("{}")
	jsonMarshallerMock.On("Marshal", expectedEvent).Return(expectedJSON, nil)

	// Ejecutar la función bajo prueba
	response, err := handler.HandleGitHubEvent(context.Background(), request)

	// Verificar resultados
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.JSONEq(t, string(expectedJSON), response.Body)

	// Verificar que se hayan llamado a los métodos esperados en los mocks
	decoderMock.AssertCalled(t, "DecodeGitHubEvent", request.Body)
	eventProcessorMock.AssertCalled(t, "ProcessEvent", mock.Anything, expectedEvent, "release")
	jsonMarshallerMock.AssertCalled(t, "Marshal", expectedEvent)
}

func TestHandleGitHubEvent_DecodeError(t *testing.T) {
	eventProcessorMock := new(MockEventProcessor)
	decoderMock := new(MockGitHubEventDecoder)
	mockLogger := new(MockLogger)
	jsonMarshallerMock := new(MockJSONMarshaler)

	handler := NewEventHandler(eventProcessorMock, mockLogger, decoderMock, jsonMarshallerMock)

	request := events.APIGatewayProxyRequest{
		Body: "invalid-json",
	}

	decoderMock.On("DecodeGitHubEvent", request.Body).Return(nil, errors.New("error al decodificar"))
	mockLogger.On("Error", "Error al decodificar el evento", mock.Anything).Return()
	response, err := handler.HandleGitHubEvent(context.Background(), request)

	assert.Error(t, err)
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	decoderMock.AssertCalled(t, "DecodeGitHubEvent", request.Body)
}

func TestHandleGitHubEvent_ProcessError(t *testing.T) {
	eventProcessorMock := new(MockEventProcessor)
	decoderMock := new(MockGitHubEventDecoder)
	loggerMock := new(MockLogger)
	jsonMarshallerMock := new(MockJSONMarshaler)

	handler := NewEventHandler(eventProcessorMock, loggerMock, decoderMock, jsonMarshallerMock)

	request := events.APIGatewayProxyRequest{
		Body: "{\"action\": \"published\"}",
	}

	expectedEvent := common.ReleaseEvent{
		Action: "published",
		Release: common.Release{
			TagName: "v1.0.0",
			Name:    "Initial Release",
		},
	}

	decoderMock.On("DecodeGitHubEvent", request.Body).Return(expectedEvent, nil)
	expectedError := errors.New("error al procesar el evento")
	eventProcessorMock.On("ProcessEvent", mock.Anything, expectedEvent, "release").Return(expectedError)
	loggerMock.On("Error", "Error al procesar el evento", mock.Anything).Once()

	response, err := handler.HandleGitHubEvent(context.Background(), request)

	assert.Error(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	assert.Contains(t, response.Body, "Error al procesar el evento: error al procesar el evento")

	decoderMock.AssertExpectations(t)
	eventProcessorMock.AssertExpectations(t)
	loggerMock.AssertExpectations(t)
}

func TestHandleGitHubEvent_JSONEncodeError(t *testing.T) {
	// Configurar mocks
	eventProcessorMock := new(MockEventProcessor)
	decoderMock := new(MockGitHubEventDecoder)
	loggerMock := new(MockLogger)
	marshalerMock := new(MockJSONMarshaler)

	// Configurar instancias de EventHandler con los mocks
	handler := NewEventHandler(eventProcessorMock, loggerMock, decoderMock, marshalerMock)

	// Configurar datos de prueba
	request := events.APIGatewayProxyRequest{
		Body: "{\"action\": \"published\"}", // Ejemplo de cuerpo de solicitud
	}

	// Simular el comportamiento del decoderMock
	expectedEvent := common.ReleaseEvent{
		Action: "published",
		Release: common.Release{
			TagName: "v1.0.0",
		},
	}
	decoderMock.On("DecodeGitHubEvent", request.Body).Return(expectedEvent, nil)
	expectedJSON := []byte{} // Definir un slice de bytes vacío o no vacío según sea necesario

	// Simular el error al codificar la respuesta a JSON
	expectedError := errors.New("error al codificar la respuesta a JSON")
	eventProcessorMock.On("ProcessEvent", mock.Anything, expectedEvent, "release").Return(nil)
	marshalerMock.On("Marshal", expectedEvent).Return(expectedJSON, expectedError)

	// Simular el registro del error
	loggerMock.On("Error", "Error al codificar la respuesta a JSON", mock.Anything).Once()

	// Ejecutar la función bajo prueba
	response, err := handler.HandleGitHubEvent(context.Background(), request)

	// Verificar resultados
	assert.Error(t, err)
	assert.NotNil(t, response) // Se espera una respuesta incluso en caso de error
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	assert.Contains(t, response.Body, "Error al codificar la respuesta a JSON")
}

func TestDecodeGitHubEvent_ReleaseEvent(t *testing.T) {
	// Configurar mocks
	loggerMock := new(MockLogger)
	decoder := NewGitHubEventDecoder(loggerMock)

	// Datos de prueba
	body := `{"action": "published", "release": {"tag_name": "V1.0"}}`
	expectedReleaseEvent := common.ReleaseEvent{Action: "published", Release: common.Release{TagName: "V1.0"}}

	// Simular el comportamiento del loggerMock
	loggerMock.On("Error", mock.Anything, mock.Anything).Once()

	// Ejecutar la función bajo prueba
	event, err := decoder.DecodeGitHubEvent(body)

	// Verificar resultados
	assert.NoError(t, err)
	assert.Equal(t, expectedReleaseEvent, event)
	loggerMock.AssertNotCalled(t, "Error", mock.Anything, mock.Anything) // Verificar que no se llamó a Logger.Error
}

func TestDecodeGitHubEvent_WorkflowJobEvent(t *testing.T) {
	// Configurar mocks
	loggerMock := new(MockLogger)
	decoder := NewGitHubEventDecoder(loggerMock)

	// Datos de prueba
	body := `{"action": "completed", "workflow_job": {"workflow_name": "CI"}}`
	expectedWorkflowEvent := common.WorkflowEvent{Action: "completed", WorkFlowJobs: common.WorkFlowJob{WorkFlowName: "CI"}}

	// Simular el comportamiento del loggerMock
	loggerMock.On("Error", mock.Anything, mock.Anything).Once()

	// Ejecutar la función bajo prueba
	event, err := decoder.DecodeGitHubEvent(body)

	// Verificar resultados
	assert.NoError(t, err)
	assert.Equal(t, expectedWorkflowEvent, event)
	loggerMock.AssertNotCalled(t, "Error", mock.Anything, mock.Anything) // Verificar que no se llamó a Logger.Error
}

// Test para DecodeGitHubEvent cuando se produce un error al decodificar el evento.
func TestDecodeGitHubEvent_DecodeError(t *testing.T) {
	// Configurar mocks
	loggerMock := new(MockLogger)
	decoder := NewGitHubEventDecoder(loggerMock)

	// Datos de prueba
	body := `{"invalid": "json"}` // JSON inválido

	// Simular el comportamiento del loggerMock
	expectedError := errors.New("acción de evento no reconocida")
	loggerMock.On("Error", "Acción de evento no reconocida", mock.Anything).Once()

	// Ejecutar la función bajo prueba
	event, err := decoder.DecodeGitHubEvent(body)

	// Verificar resultados
	assert.Error(t, err)
	assert.Nil(t, event)
	assert.EqualError(t, err, expectedError.Error())
	loggerMock.AssertCalled(t, "Error", "Acción de evento no reconocida", mock.Anything) // Verificar que se llamó a Logger.Error
}
