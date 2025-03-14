//go:build !integration

package service

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestTopicPublisherService_PublishMediaProcessed(t *testing.T) {
	mockMessageQueue := new(MockMessageQueue)
	mockLogger := new(logger.MockLogger)

	service := NewMediaProcessingPublisherService(mockMessageQueue, mockLogger)

	ctx := context.Background()
	message := &model.MediaProcessingMessage{
		VideoID: "video123",
		Status:  "processed",
		Message: "success",
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockMessageQueue.On("SendMessage", ctx, message).Return(nil)

	// Act
	err := service.PublishMediaProcessed(ctx, message)

	// Assert
	assert.NoError(t, err, "No se esperaba un error al publicar el mensaje")
	mockMessageQueue.AssertCalled(t, "SendMessage", ctx, message)
}

func TestTopicPublisherService_PublishMediaProcessed_Error(t *testing.T) {
	// Arrange
	mockMessageQueue := new(MockMessageQueue)
	mockLogger := new(logger.MockLogger)

	service := NewMediaProcessingPublisherService(mockMessageQueue, mockLogger)

	ctx := context.Background()
	message := &model.MediaProcessingMessage{
		VideoID: "video123",
		Status:  "processed",
		Message: "success",
	}

	expectedError := errors.New("error al enviar el mensaje")
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockMessageQueue.On("SendMessage", ctx, message).Return(expectedError)

	// Act
	err := service.PublishMediaProcessed(ctx, message)

	// Assert
	assert.Error(t, err, "Se esperaba un error al publicar el mensaje")
	assert.Contains(t, err.Error(), expectedError.Error(), "El mensaje de error no es el esperado")
	mockMessageQueue.AssertCalled(t, "SendMessage", ctx, message)
}
