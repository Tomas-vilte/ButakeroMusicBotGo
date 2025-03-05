//go:build !integration

package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestMessagingService_SendProcessingMessage(t *testing.T) {
	t.Run("should send processing message successfully", func(t *testing.T) {
		// Arrange
		mockMessagingQueue := new(MockMessagingQueue)
		mockLogger := new(logger.MockLogger)

		messagingService := NewMessagingService(mockMessagingQueue, mockLogger)

		ctx := context.Background()
		operationID := "test_operation_id"
		status := "success"
		metadata := &model.Metadata{
			VideoID: "test_video_id",
		}
		attempts := 1

		mockMessagingQueue.On("SendMessage", ctx, mock.MatchedBy(func(message model.Message) bool {
			return message.ID == operationID &&
				message.Content == fmt.Sprintf("Procesamiento %s", status) &&
				message.Status.ID == operationID &&
				message.Status.SK == metadata.VideoID &&
				message.Status.Status == status &&
				message.Status.Metadata == metadata &&
				message.Status.Attempts == attempts
		})).Return(nil)
		// Act
		err := messagingService.SendProcessingMessage(ctx, operationID, status, metadata, attempts)

		// Assert
		assert.NoError(t, err)
		mockMessagingQueue.AssertExpectations(t)
	})

	t.Run("should return error when sending message fails", func(t *testing.T) {
		// Arrange
		mockMessagingQueue := new(MockMessagingQueue)
		mockLogger := new(logger.MockLogger)

		messagingService := NewMessagingService(mockMessagingQueue, mockLogger)

		ctx := context.Background()
		operationID := "test_operation_id"
		status := "success"
		metadata := &model.Metadata{
			VideoID: "test_video_id",
		}
		attempts := 1

		mockMessagingQueue.On("SendMessage", ctx, mock.MatchedBy(func(message model.Message) bool {
			return message.ID == operationID &&
				message.Content == fmt.Sprintf("Procesamiento %s", status) &&
				message.Status.ID == operationID &&
				message.Status.SK == metadata.VideoID &&
				message.Status.Status == status &&
				message.Status.Metadata == metadata &&
				message.Status.Attempts == attempts
		})).Return(errors.New("send failed"))
		// Act
		err := messagingService.SendProcessingMessage(ctx, operationID, status, metadata, attempts)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Error al enviar mensaje")
		mockMessagingQueue.AssertExpectations(t)
	})
}
