//go:build !integration

package service

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestVideoService_GetMediaDetails(t *testing.T) {
	t.Run("should return media details successfully", func(t *testing.T) {
		// arrange
		mockProvider := new(MockVideoProvider)
		mockLogger := new(logger.MockLogger)

		providers := map[string]ports.VideoProvider{
			"youtube": mockProvider,
		}

		videoService := NewVideoService(providers, mockLogger)

		ctx := context.Background()
		input := "test_video"
		providerType := "youtube"
		videoID := "12345"
		mediaDetails := &model.MediaDetails{
			Title:       "Test Video",
			Description: "This is a test video",
			DurationMs:  257026,
		}

		mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
		mockProvider.On("SearchVideoID", ctx, input).Return(videoID, nil)
		mockProvider.On("GetVideoDetails", ctx, videoID).Return(mediaDetails, nil)

		// act
		result, err := videoService.GetMediaDetails(ctx, input, providerType)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, mediaDetails, result)
		mockProvider.AssertExpectations(t)

	})

	t.Run("should return error when searching video ID fails", func(t *testing.T) {
		// Arrange
		mockProvider := new(MockVideoProvider)
		mockLogger := new(logger.MockLogger)

		providers := map[string]ports.VideoProvider{
			"youtube": mockProvider,
		}

		videoService := NewVideoService(providers, mockLogger)

		ctx := context.Background()
		input := "test_video"
		providerType := "youtube"

		mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()
		mockProvider.On("SearchVideoID", ctx, input).Return("", errors.New("search failed"))

		// Act
		result, err := videoService.GetMediaDetails(ctx, input, providerType)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Error al buscar ID de video")
		mockProvider.AssertExpectations(t)
	})

	t.Run("should return error when searching video ID fails", func(t *testing.T) {
		// Arrange
		mockProvider := new(MockVideoProvider)
		mockLogger := new(logger.MockLogger)

		providers := map[string]ports.VideoProvider{
			"youtube": mockProvider,
		}

		videoService := NewVideoService(providers, mockLogger)

		ctx := context.Background()
		input := "test_video"
		providerType := "youtube"

		mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()
		mockProvider.On("SearchVideoID", ctx, input).Return("", errors.New("search failed"))

		// Act
		result, err := videoService.GetMediaDetails(ctx, input, providerType)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Error al buscar ID de video")
		mockProvider.AssertExpectations(t)
	})

	t.Run("should return error when getting video details fails", func(t *testing.T) {
		// Arrange
		mockProvider := new(MockVideoProvider)
		mockLogger := new(logger.MockLogger)

		providers := map[string]ports.VideoProvider{
			"youtube": mockProvider,
		}

		videoService := NewVideoService(providers, mockLogger)

		ctx := context.Background()
		input := "test_video"
		providerType := "youtube"
		videoID := "12345"

		mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()
		mockProvider.On("SearchVideoID", ctx, input).Return(videoID, nil)
		mockProvider.On("GetVideoDetails", ctx, videoID).Return(&model.MediaDetails{}, errors.New("details failed"))

		// Act
		result, err := videoService.GetMediaDetails(ctx, input, providerType)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "Error al obtener detalles de la cancion")
		mockProvider.AssertExpectations(t)
	})
}
