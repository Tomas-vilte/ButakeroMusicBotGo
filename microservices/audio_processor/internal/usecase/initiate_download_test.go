//go:build !integration

package usecase

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
)

func TestInitiateDownloadUseCase_Execute(t *testing.T) {
	t.Run("It should initiate download successfully", func(t *testing.T) {
		// Arrange
		mockCoreService := new(MockCoreService)
		mockVideoService := new(MockVideoService)
		mockOperationService := new(MockOperationService)

		uc := NewInitiateDownloadUseCase(mockCoreService, mockVideoService, mockOperationService)

		ctx := context.Background()
		song := "test-song"
		providerType := "youtube"
		mediaDetails := &model.MediaDetails{
			ID:    "test-media-id",
			Title: "Test Song",
		}
		operationResult := &model.OperationInitResult{
			Status: "started",
		}

		var wg sync.WaitGroup
		wg.Add(1)

		mockVideoService.On("GetMediaDetails", ctx, song, providerType).Return(mediaDetails, nil)
		mockOperationService.On("StartOperation", ctx, mediaDetails).Return(operationResult, nil)
		mockCoreService.On("ProcessMedia", context.Background(), mediaDetails).Run(func(args mock.Arguments) {
			wg.Done()
		}).Return(nil)

		// Act
		result, err := uc.Execute(ctx, song, providerType)

		wg.Wait()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, operationResult, result)
		mockVideoService.AssertExpectations(t)
		mockOperationService.AssertExpectations(t)
		mockCoreService.AssertExpectations(t)
	})

	t.Run("It should handle background processing error", func(t *testing.T) {
		// Arrange
		mockCoreService := new(MockCoreService)
		mockVideoService := new(MockVideoService)
		mockOperationService := new(MockOperationService)

		uc := NewInitiateDownloadUseCase(mockCoreService, mockVideoService, mockOperationService)

		ctx := context.Background()
		song := "test-song"
		providerType := "youtube"
		mediaDetails := &model.MediaDetails{
			ID:    "test-media-id",
			Title: "Test Song",
		}
		operationResult := &model.OperationInitResult{
			Status: "started",
		}
		expectedError := errors.New("background processing failed")

		var wg sync.WaitGroup
		wg.Add(1)

		mockVideoService.On("GetMediaDetails", ctx, song, providerType).Return(mediaDetails, nil)
		mockOperationService.On("StartOperation", ctx, mediaDetails).Return(operationResult, nil)
		mockCoreService.On("ProcessMedia", context.Background(), mediaDetails).Run(func(args mock.Arguments) {
			wg.Done()
		}).Return(expectedError)

		// Act
		result, err := uc.Execute(ctx, song, providerType)

		wg.Wait()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, operationResult, result)
		mockVideoService.AssertExpectations(t)
		mockOperationService.AssertExpectations(t)
		mockCoreService.AssertExpectations(t)
	})
}
