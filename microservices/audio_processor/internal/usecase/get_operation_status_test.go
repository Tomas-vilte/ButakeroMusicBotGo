//go:build !integration

package usecase

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetOperationStatusUseCase_Execute(t *testing.T) {

	t.Run("It should return the operation when it is found", func(t *testing.T) {
		mockMediaRepo := new(MockMediaRepository)
		ctx := context.Background()
		uc := NewGetOperationStatusUseCase(mockMediaRepo)

		videoID := "test-video-id"
		expectedOperation := &model.Media{
			Status: "completed",
		}

		mockMediaRepo.On("GetMedia", ctx, videoID).Return(expectedOperation, nil)

		result, err := uc.Execute(ctx, "test-video-id")

		assert.NoError(t, err)
		assert.Equal(t, expectedOperation, result)
		mockMediaRepo.AssertExpectations(t)
	})

	t.Run("It should return an error when the operation is not found", func(t *testing.T) {
		mockMediaRepo := new(MockMediaRepository)
		ctx := context.Background()
		uc := NewGetOperationStatusUseCase(mockMediaRepo)

		videoID := "test-video-id"

		mockMediaRepo.On("GetMedia", ctx, videoID).Return((*model.Media)(nil), errors.ErrOperationNotFound)

		result, err := uc.Execute(ctx, videoID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, errors.ErrOperationNotFound, err)
		mockMediaRepo.AssertExpectations(t)
	})

	t.Run("It should return an error when the repository returns an error", func(t *testing.T) {
		mockMediaRepo := new(MockMediaRepository)
		ctx := context.Background()
		uc := NewGetOperationStatusUseCase(mockMediaRepo)

		videoID := "test-video-id"

		mockMediaRepo.On("GetMedia", ctx, videoID).Return((*model.Media)(nil), fmt.Errorf("unexpected error"))

		result, err := uc.Execute(ctx, videoID)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockMediaRepo.AssertExpectations(t)
	})

	t.Run("It should return an error when the context is canceled", func(t *testing.T) {
		mockMediaRepo := new(MockMediaRepository)
		ctx, cancel := context.WithCancel(context.Background())
		uc := NewGetOperationStatusUseCase(mockMediaRepo)

		videoID := "test-video-id"

		cancel()

		mockMediaRepo.On("GetMedia", ctx, videoID).Return((*model.Media)(nil), context.Canceled)

		result, err := uc.Execute(ctx, videoID)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockMediaRepo.AssertExpectations(t)
	})

	t.Run("It should return an error when the song ID is empty", func(t *testing.T) {
		mockMediaRepo := new(MockMediaRepository)
		ctx := context.Background()
		uc := NewGetOperationStatusUseCase(mockMediaRepo)

		videoID := ""

		result, err := uc.Execute(ctx, videoID)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockMediaRepo.AssertExpectations(t)
	})

}
