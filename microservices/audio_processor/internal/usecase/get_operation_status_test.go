//go:build !integration

package usecase

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetOperationStatusUseCase_Execute(t *testing.T) {

	t.Run("It should return the operation when it is found", func(t *testing.T) {
		mockMediaRepo := new(MockMediaRepository)
		ctx := context.Background()
		uc := NewGetOperationStatusUseCase(mockMediaRepo)

		operationID := uuid.New().String()
		songID := "test-song-id"
		expectedOperation := &model.Media{
			ID:     operationID,
			Status: "completed",
		}

		mockMediaRepo.On("GetMedia", ctx, operationID, songID).Return(expectedOperation, nil)

		result, err := uc.Execute(ctx, operationID, "test-song-id")

		assert.NoError(t, err)
		assert.Equal(t, expectedOperation, result)
		mockMediaRepo.AssertExpectations(t)
	})

}
