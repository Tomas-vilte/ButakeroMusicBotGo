package service

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"testing"
)

type MockPublisher struct {
	mock.Mock
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

func (m *MockPublisher) Publish(ctx context.Context, event interface{}, eventType string) error {
	args := m.Called(ctx, event, eventType)
	return args.Error(0)
}

func TestProcessEvent_Success(t *testing.T) {
	mockPublisher := &MockPublisher{}
	mockLogger := &MockLogger{}

	releaseEvent := common.ReleaseEvent{
		Action: "published",
		Release: common.Release{
			TagName: "v1.0.0",
			Name:    "Initial Release",
		},
	}

	mockPublisher.On("Publish", mock.Anything, releaseEvent, "published").Return(nil)
	mockLogger.On("Info", "Evento publicado en SQS con exito", mock.Anything).Once()

	eventProcessor := NewEventProcessor(mockPublisher, mockLogger)
	err := eventProcessor.ProcessEvent(context.Background(), releaseEvent, "published")

	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestProcessEvent_PublishError(t *testing.T) {
	mockPublisher := &MockPublisher{}
	mockLogger := &MockLogger{}

	releaseEvent := common.ReleaseEvent{
		Action: "published",
		Release: common.Release{
			TagName: "v1.0.0",
			Name:    "Initial Release",
		},
	}

	expectedError := errors.New("error al publicar")
	mockPublisher.On("Publish", mock.Anything, releaseEvent, "published").Return(expectedError)
	mockLogger.On("Error", "Error publicando el evento a SQS", mock.Anything).Once()

	eventProcessor := NewEventProcessor(mockPublisher, mockLogger)
	err := eventProcessor.ProcessEvent(context.Background(), releaseEvent, "published")

	assert.Equal(t, expectedError, err)
	mockPublisher.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestProcessEvent_ValidationError(t *testing.T) {
	mockPublisher := new(MockPublisher)
	mockLogger := new(MockLogger)

	event := common.ReleaseEvent{
		Action: "published",
		Release: common.Release{
			TagName: "v1.0.0",
			Name:    "", // Invalid because Name is empty
		},
	}

	mockLogger.On("Error", "Error de validación del evento", mock.Anything).Once()

	processor := NewEventProcessor(mockPublisher, mockLogger)
	err := processor.ProcessEvent(context.Background(), event, "published")

	assert.EqualError(t, err, "el campo 'Name' en el evento de lanzamiento está vacío")
	mockPublisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
	mockLogger.AssertExpectations(t)
}

func TestProcessEvent_UnsupportedEventType(t *testing.T) {
	mockPublisher := new(MockPublisher)
	mockLogger := new(MockLogger)

	invalidEvent := struct {
		Action string
	}{Action: "invalid"}

	mockLogger.On("Error", "Error de validación del evento", mock.Anything).Once()

	processor := NewEventProcessor(mockPublisher, mockLogger)
	err := processor.ProcessEvent(context.Background(), invalidEvent, "invalid")

	assert.EqualError(t, err, "tipo de evento no compatible para validación")
	mockPublisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
	mockLogger.AssertExpectations(t)
}

func TestProcessEvent_WorkflowEventValidationError(t *testing.T) {
	mockPublisher := new(MockPublisher)
	mockLogger := new(MockLogger)

	event := common.WorkflowEvent{
		Action: "started",
		WorkFlowJobs: common.WorkFlowJob{
			WorkFlowName: "",
			ID:           23565,
		},
	}

	mockLogger.On("Error", "Error de validación del evento", mock.Anything).Once()

	processor := NewEventProcessor(mockPublisher, mockLogger)
	err := processor.ProcessEvent(context.Background(), event, "started")

	assert.EqualError(t, err, "el campo 'Name' en el evento de flujo de trabajo está vacío")
	mockPublisher.AssertNotCalled(t, "Publish", mock.Anything, mock.Anything, mock.Anything)
	mockLogger.AssertExpectations(t)
}
