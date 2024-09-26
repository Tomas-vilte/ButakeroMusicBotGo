package unit

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/usecase"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetOperationStatusUseCase_Execute(t *testing.T) {

	t.Run("It should return the operation when it is found", func(t *testing.T) {
		mockRepo := new(MockOperationRepository)
		ctx := context.Background()
		uc := usecase.NewGetOperationStatusUseCase(mockRepo)
		expectedOperation := &model.OperationResult{
			ID:     "operation-id",
			Status: "completed",
		}

		mockRepo.On("GetOperationResult", ctx, "operation-id", "song-id").Return(expectedOperation, nil)

		result, err := uc.Execute(ctx, "operation-id", "song-id")

		assert.NoError(t, err)
		assert.Equal(t, expectedOperation, &result)
		mockRepo.AssertExpectations(t)
	})

}
