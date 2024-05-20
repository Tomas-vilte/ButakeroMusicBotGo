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

func (m *MockPublisher) Publish(ctx context.Context, event interface{}) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func TestProcessEvent_Success(t *testing.T) {
	mockPublisher := &MockPublisher{}
	mockLogger := &MockLogger{}

	event := common.Event{
		Release: common.Release{
			TagName: "v1.0.0",
		},
	}

	mockPublisher.On("Publish", mock.Anything, event).Return(nil)
	mockLogger.On("Info", "Procesando evento", []zap.Field{
		zap.String("Tag", event.Release.TagName),
	}).Return()
	mockLogger.On("Info", "Evento publicado en SQS con exito", mock.Anything).Return()

	processor := NewEventProcessor(mockPublisher, mockLogger)
	err := processor.ProcessEvent(context.Background(), event)

	assert.Nil(t, err)
	mockPublisher.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestProcessEvent_PublishError(t *testing.T) {
	mockPublisher := &MockPublisher{}
	mockLogger := &MockLogger{}

	event := common.Event{
		Release: common.Release{
			TagName: "v1.0.0",
		},
	}

	expectedError := errors.New("publish error")
	mockLogger.On("Info", "Procesando evento", []zap.Field{
		zap.String("Tag", event.Release.TagName),
	}).Return()
	mockPublisher.On("Publish", mock.Anything, event).Return(expectedError)

	mockLogger.On("Error", "Error publicando el evento a SQS", mock.Anything).Return()

	processor := NewEventProcessor(mockPublisher, mockLogger)
	err := processor.ProcessEvent(context.Background(), event)

	assert.Equal(t, expectedError, err)
	mockPublisher.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
