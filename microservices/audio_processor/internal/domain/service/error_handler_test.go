//go:build !integration

package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestErrorHandler_HandleProcessingError(t *testing.T) {
	t.Run("should handle processing error successfully", func(t *testing.T) {
		// Arrange
		mockOpsRepo := new(MockOperationRepository)
		mockMessaging := new(MockMessagingQueue)
		mockLogger := new(logger.MockLogger)
		cfg := &config.Config{}

		errorHandler := NewErrorHandler(mockOpsRepo, mockMessaging, mockLogger, cfg)

		ctx := context.Background()
		operationID := "test_operation_id"
		metadata := &model.Metadata{
			VideoID: "test_video_id",
		}
		errorType := "test_error_type"
		attempts := 1
		originalErr := errors.New("test error")

		expectedResult := &model.OperationResult{
			ID:             operationID,
			SK:             metadata.VideoID,
			Status:         statusFailed,
			Message:        fmt.Sprintf("[%s] %v", errorType, originalErr),
			Metadata:       metadata,
			ProcessingDate: time.Now().Format(time.RFC3339),
			Attempts:       attempts,
			Failures:       attempts,
		}

		expectedMessage := model.Message{
			ID:      operationID,
			Content: "Error en procesamiento",
			Status: model.Status{
				ID:             operationID,
				SK:             metadata.VideoID,
				Status:         statusFailed,
				Message:        expectedResult.Message,
				Metadata:       metadata,
				ProcessingDate: time.Now().UTC(),
				Attempts:       attempts,
				Failures:       attempts,
			},
		}

		mockOpsRepo.On("UpdateOperationResult", ctx, operationID, mock.MatchedBy(func(result *model.OperationResult) bool {
			return result.ID == expectedResult.ID &&
				result.SK == expectedResult.SK &&
				result.Status == expectedResult.Status &&
				result.Message == expectedResult.Message &&
				result.Metadata == expectedResult.Metadata &&
				result.Attempts == expectedResult.Attempts &&
				result.Failures == expectedResult.Failures
		})).Return(nil)

		mockMessaging.On("SendMessage", ctx, mock.MatchedBy(func(message model.Message) bool {
			return message.ID == expectedMessage.ID &&
				message.Content == expectedMessage.Content &&
				message.Status.ID == expectedMessage.Status.ID &&
				message.Status.SK == expectedMessage.Status.SK &&
				message.Status.Status == expectedMessage.Status.Status &&
				message.Status.Message == expectedMessage.Status.Message &&
				message.Status.Metadata == expectedMessage.Status.Metadata &&
				message.Status.Attempts == expectedMessage.Status.Attempts &&
				message.Status.Failures == expectedMessage.Status.Failures
		})).Return(nil)

		// Act
		err := errorHandler.HandleProcessingError(ctx, operationID, metadata, errorType, attempts, originalErr)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, originalErr, err)
		mockOpsRepo.AssertExpectations(t)
		mockMessaging.AssertExpectations(t)
	})

	t.Run("should log error when updating operation result fails", func(t *testing.T) {
		// Arrange
		mockOpsRepo := new(MockOperationRepository)
		mockMessaging := new(MockMessagingQueue)
		mockLogger := new(logger.MockLogger)
		cfg := &config.Config{}
		errorHandler := NewErrorHandler(mockOpsRepo, mockMessaging, mockLogger, cfg)

		ctx := context.Background()
		operationID := "test_operation_id"
		metadata := &model.Metadata{
			VideoID: "test_video_id",
		}
		errorType := "test_error_type"
		attempts := 1
		originalErr := errors.New("test error")

		expectedResult := &model.OperationResult{
			ID:             operationID,
			SK:             metadata.VideoID,
			Status:         statusFailed,
			Message:        fmt.Sprintf("[%s] %v", errorType, originalErr),
			Metadata:       metadata,
			ProcessingDate: time.Now().Format(time.RFC3339),
			Attempts:       attempts,
			Failures:       attempts,
		}

		mockOpsRepo.On("UpdateOperationResult", ctx, operationID, mock.MatchedBy(func(result *model.OperationResult) bool {
			return result.ID == expectedResult.ID &&
				result.SK == expectedResult.SK &&
				result.Status == expectedResult.Status &&
				result.Message == expectedResult.Message &&
				result.Metadata == expectedResult.Metadata &&
				result.Attempts == expectedResult.Attempts &&
				result.Failures == expectedResult.Failures
		})).Return(errors.New("update failed"))
		mockMessaging.On("SendMessage", ctx, mock.Anything).Return(nil)

		mockLogger.On("Error", "Error al actualizar estado fallido", mock.Anything).Return()

		// Act
		err := errorHandler.HandleProcessingError(ctx, operationID, metadata, errorType, attempts, originalErr)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, originalErr, err)
		mockOpsRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("should log error when sending message fails", func(t *testing.T) {
		// Arrange
		mockOpsRepo := new(MockOperationRepository)
		mockMessaging := new(MockMessagingQueue)
		mockLogger := new(logger.MockLogger)
		cfg := &config.Config{}

		errorHandler := NewErrorHandler(mockOpsRepo, mockMessaging, mockLogger, cfg)

		ctx := context.Background()
		operationID := "test_operation_id"
		metadata := &model.Metadata{
			VideoID: "test_video_id",
		}
		errorType := "test_error_type"
		attempts := 1
		originalErr := errors.New("test error")

		expectedResult := &model.OperationResult{
			ID:             operationID,
			SK:             metadata.VideoID,
			Status:         statusFailed,
			Message:        fmt.Sprintf("[%s] %v", errorType, originalErr),
			Metadata:       metadata,
			ProcessingDate: time.Now().Format(time.RFC3339),
			Attempts:       attempts,
			Failures:       attempts,
		}

		expectedMessage := model.Message{
			ID:      operationID,
			Content: "Error en procesamiento",
			Status: model.Status{
				ID:             operationID,
				SK:             metadata.VideoID,
				Status:         statusFailed,
				Message:        expectedResult.Message,
				Metadata:       metadata,
				ProcessingDate: time.Now().UTC(),
				Attempts:       attempts,
				Failures:       attempts,
			},
		}

		mockOpsRepo.On("UpdateOperationResult", ctx, operationID, mock.MatchedBy(func(result *model.OperationResult) bool {
			return result.ID == expectedResult.ID &&
				result.SK == expectedResult.SK &&
				result.Status == expectedResult.Status &&
				result.Message == expectedResult.Message &&
				result.Metadata == expectedResult.Metadata &&
				result.Attempts == expectedResult.Attempts &&
				result.Failures == expectedResult.Failures
		})).Return(nil)

		mockMessaging.On("SendMessage", ctx, mock.MatchedBy(func(message model.Message) bool {
			return message.ID == expectedMessage.ID &&
				message.Content == expectedMessage.Content &&
				message.Status.ID == expectedMessage.Status.ID &&
				message.Status.SK == expectedMessage.Status.SK &&
				message.Status.Status == expectedMessage.Status.Status &&
				message.Status.Message == expectedMessage.Status.Message &&
				message.Status.Metadata == expectedMessage.Status.Metadata &&
				message.Status.Attempts == expectedMessage.Status.Attempts &&
				message.Status.Failures == expectedMessage.Status.Failures
		})).Return(errors.New("send failed"))

		mockLogger.On("Error", "Error al enviar mensaje de fallo", mock.Anything).Return()

		// Act
		err := errorHandler.HandleProcessingError(ctx, operationID, metadata, errorType, attempts, originalErr)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, originalErr, err)
		mockOpsRepo.AssertExpectations(t)
		mockMessaging.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
