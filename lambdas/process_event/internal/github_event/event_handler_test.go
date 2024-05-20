package github_event

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/common"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"testing"
)

type MockEventProcessor struct {
	mock.Mock
}

func (m *MockEventProcessor) ProcessEvent(ctx context.Context, event common.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func TestHandleGitHubEvent_Success(t *testing.T) {
	mockEventProcessor := &MockEventProcessor{}
	mockLogger := &MockLogger{}

	event := common.Event{
		Action: "test_action",
		Release: common.Release{
			TagName: "v1.0.0",
		},
	}
	eventJSON, _ := json.Marshal(event)

	mockEventProcessor.On("ProcessEvent", mock.Anything, event).Return(nil)
	mockLogger.On("Info", "Evento recibido de Github", []zapcore.Field{
		zap.String("Action", event.Action),
		zap.String("Tag", event.Release.TagName),
	}).Return()

	handler := NewEventHandler(mockEventProcessor, mockLogger)
	request := events.APIGatewayProxyRequest{Body: string(eventJSON)}

	response, err := handler.HandleGitHubEvent(context.Background(), request)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	mockEventProcessor.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestHandleGitHubEvent_UnmarshalError(t *testing.T) {
	mockProcessor := &MockEventProcessor{}
	mockLogger := &MockLogger{}

	mockLogger.On("Error", "Error al decodificar el evento", mock.Anything).Return()

	handler := NewEventHandler(mockProcessor, mockLogger)
	request := events.APIGatewayProxyRequest{Body: "invalid json"}

	response, err := handler.HandleGitHubEvent(context.Background(), request)

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	mockProcessor.AssertNotCalled(t, "ProcessEvent", mock.Anything, mock.Anything)
	mockLogger.AssertExpectations(t)
}

func TestHandleGitHubEvent_ProcessEventError(t *testing.T) {
	mockProcessor := new(MockEventProcessor)
	mockLogger := new(MockLogger)

	event := common.Event{
		Action: "test_action",
		Release: common.Release{
			TagName: "v1.0.0",
		},
	}
	eventJSON, _ := json.Marshal(event)

	mockProcessor.On("ProcessEvent", mock.Anything, event).Return(errors.New("processing error"))
	mockLogger.On("Info", "Evento recibido de Github", []zapcore.Field{
		zap.String("Action", event.Action),
		zap.String("Tag", event.Release.TagName),
	}).Return()
	mockLogger.On("Error", "Error al procesar evento", mock.Anything).Return()

	handler := NewEventHandler(mockProcessor, mockLogger)
	request := events.APIGatewayProxyRequest{Body: string(eventJSON)}

	response, err := handler.HandleGitHubEvent(context.Background(), request)

	assert.NotNil(t, err)
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
	mockProcessor.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
