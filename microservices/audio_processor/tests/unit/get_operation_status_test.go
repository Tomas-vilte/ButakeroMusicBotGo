package unit

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetOperationStatusUseCase_Execute(t *testing.T) {

	t.Run("It should return the operation when it is found", func(t *testing.T) {
		mockRepo := new(MockOperationRepository)
		ctx := context.Background()
		uc := usecase.NewGetOperationStatusUseCase(mockRepo)

		operationID := uuid.New().String()
		expectedOperation := &model.OperationResult{
			PK:     operationID,
			Status: "completed",
		}

		mockRepo.On("GetOperationResult", ctx, operationID, "song-id").Return(expectedOperation, nil)

		result, err := uc.Execute(ctx, operationID, "song-id")

		assert.NoError(t, err)
		assert.Equal(t, expectedOperation, result)
		mockRepo.AssertExpectations(t)
	})

}
