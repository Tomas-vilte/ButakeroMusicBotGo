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

func TestOperationService_StartOperation(t *testing.T) {
	t.Run("should start operation successfully", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockOperationRepository)
		mockLogger := new(logger.MockLogger)

		operationService := NewOperationService(mockRepo, mockLogger)

		ctx := context.Background()
		songID := "test_song_id"

		mockRepo.On("SaveOperationsResult", ctx, mock.MatchedBy(func(operation *model.OperationResult) bool {
			return operation.SK == songID && operation.Status == statusInitiating
		})).Return(nil)

		// Act
		result, err := operationService.StartOperation(ctx, songID)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, songID, result.SongID)
		assert.Equal(t, statusInitiating, result.Status)
		assert.NotEmpty(t, result.CreatedAt)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when saving operation fails", func(t *testing.T) {
		// Arrange
		mockRepo := new(MockOperationRepository)
		mockLogger := new(logger.MockLogger)

		operationService := NewOperationService(mockRepo, mockLogger)

		ctx := context.Background()
		songID := "test_song_id"

		mockLogger.On("Error", mock.Anything, mock.Anything).Return()
		mockRepo.On("SaveOperationsResult", ctx, mock.MatchedBy(func(operation *model.OperationResult) bool {
			return operation.SK == songID && operation.Status == statusInitiating
		})).Return(errors.New("save failed"))

		// Act
		result, err := operationService.StartOperation(ctx, songID)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Error al iniciar la operación")
		mockRepo.AssertExpectations(t)
	})
}
