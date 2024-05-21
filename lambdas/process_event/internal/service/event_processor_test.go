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
	// Configurar mocks
	mockPublisher := new(MockPublisher)
	mockLogger := new(MockLogger)

	// Datos de prueba
	event := common.Event{
		Release: common.Release{
			TagName: "v1.0.0",
		},
	}

	// Simular el comportamiento del publisher
	mockPublisher.On("Publish", mock.Anything, event).Return(nil)

	// Simular el comportamiento del logger
	mockLogger.On("Info", "Evento publicado en SQS con exito", mock.Anything).Once()

	// Ejecutar la función bajo prueba
	processor := NewEventProcessor(mockPublisher, mockLogger)
	err := processor.ProcessEvent(context.Background(), event)

	// Verificar resultados
	assert.NoError(t, err)
	mockPublisher.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestProcessEvent_PublishError(t *testing.T) {
	// Configurar mocks
	mockPublisher := new(MockPublisher)
	mockLogger := new(MockLogger)

	// Datos de prueba
	event := common.Event{
		Release: common.Release{
			TagName: "v1.0.0",
		},
	}

	// Simular el comportamiento del publisher
	expectedError := errors.New("error al publicar")
	mockPublisher.On("Publish", mock.Anything, event).Return(expectedError)

	// Simular el comportamiento del logger
	mockLogger.On("Error", "Error publicando el evento a SQS", mock.Anything).Once()

	// Ejecutar la función bajo prueba
	processor := NewEventProcessor(mockPublisher, mockLogger)
	err := processor.ProcessEvent(context.Background(), event)

	// Verificar resultados
	assert.Equal(t, expectedError, err)
	mockPublisher.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
