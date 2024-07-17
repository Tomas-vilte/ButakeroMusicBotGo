package lambda

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestHandler_Handle_Success(t *testing.T) {
	mockUseCase := new(MockSendJobToECS)
	mockLogger := new(MockLogger)
	handler := NewHandler(mockUseCase, mockLogger)

	event := events.S3Event{
		Records: []events.S3EventRecord{
			{
				S3: events.S3Entity{
					Bucket: events.S3Bucket{Name: "test-bucket"},
					Object: events.S3Object{Key: "test-key"},
				},
				AWSRegion: "us-east-1",
			},
		},
	}

	mockUseCase.On("Execute", mock.Anything, mock.AnythingOfType("entity.Job")).Return(nil)
	mockLogger.On("Info", "Procesando evento de audio", mock.Anything).Return(nil)

	err := handler.Handle(context.Background(), event)

	assert.NoError(t, err)
	mockUseCase.AssertExpectations(t)
	mockLogger.AssertExpectations(t)

}

func TestHandler_Handle_Error(t *testing.T) {
	mockUseCase := new(MockSendJobToECS)
	mockLogger := new(MockLogger)
	handler := NewHandler(mockUseCase, mockLogger)

	event := events.S3Event{
		Records: []events.S3EventRecord{
			{
				S3: events.S3Entity{
					Bucket: events.S3Bucket{
						Name: "test-bucket",
					},
					Object: events.S3Object{
						Key: "test-audio-file.mp3",
					},
				},
				AWSRegion: "us-east-1",
			},
		},
	}

	mockUseCase.On("Execute", mock.Anything, mock.AnythingOfType("entity.Job")).Return(errors.New("execution error"))
	mockLogger.On("Info", "Procesando evento de audio", mock.Anything).Return(nil)
	mockLogger.On("Error", "Error al enviar el job a ECS", mock.Anything).Return(nil)

	err := handler.Handle(context.Background(), event)

	assert.Error(t, err)
	assert.EqualError(t, err, "execution error")
	mockUseCase.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
