//go:build !integration

package service

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestOperationManager_HandleOperationSuccess(t *testing.T) {
	t.Run("should handle operation success successfully", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockOperationRepository)
		mockLogger := new(logger.MockLogger)
		cfg := &config.Config{}

		operationManager := NewOperationManager(mockRepo, mockLogger, cfg)

		ctx := context.Background()
		operationID := "test_operation_id"
		metadata := &model.Metadata{
			VideoID: "test_video_id",
		}
		fileData := &model.FileData{
			FilePath: "test_key",
			FileSize: "1024",
		}

		expectedResult := &model.OperationResult{
			ID:             operationID,
			SK:             metadata.VideoID,
			Status:         statusSuccess,
			Message:        "Procesamiento completado exitosamente",
			Metadata:       metadata,
			FileData:       fileData,
			ProcessingDate: time.Now().Format(time.RFC3339),
			Success:        true,
			Attempts:       1,
			Failures:       0,
		}

		mockRepo.On("UpdateOperationResult", ctx, operationID, expectedResult).Return(nil)

		// Act
		err := operationManager.HandleOperationSuccess(ctx, operationID, metadata, fileData)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when updating operation result fails", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockOperationRepository)
		mockLogger := new(logger.MockLogger)
		cfg := &config.Config{}

		operationManager := NewOperationManager(mockRepo, mockLogger, cfg)

		ctx := context.Background()
		operationID := "test_operation_id"
		metadata := &model.Metadata{
			VideoID: "test_video_id",
		}
		fileData := &model.FileData{
			FilePath: "test_key",
			FileSize: "1024",
		}

		expectedResult := &model.OperationResult{
			ID:             operationID,
			SK:             metadata.VideoID,
			Status:         statusSuccess,
			Message:        "Procesamiento completado exitosamente",
			Metadata:       metadata,
			FileData:       fileData,
			ProcessingDate: time.Now().Format(time.RFC3339),
			Success:        true,
			Attempts:       1,
			Failures:       0,
		}

		mockRepo.On("UpdateOperationResult", ctx, operationID, expectedResult).Return(errors.New("update failed"))

		// Act
		err := operationManager.HandleOperationSuccess(ctx, operationID, metadata, fileData)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update failed")
		mockRepo.AssertExpectations(t)
	})
}
