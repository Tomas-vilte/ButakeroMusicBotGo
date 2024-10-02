package unit

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestInitiateDownloadUseCase_Execute(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		mockYouTubeService := new(MockYouTubeService)
		mockAudioService := new(MockAudioProcessingService)

		uc := usecase.NewInitiateDownloadUseCase(mockAudioService, mockYouTubeService)

		ctx := context.Background()
		song := "test song"

		// Setup expectativas
		videoID := "test-video-id"
		youtubeMetadata := &api.VideoDetails{
			VideoID:    videoID,
			Title:      "Test Video",
			Duration:   "3:00",
			URLYouTube: "https://youtube.com/watch?v=test-video-id",
		}

		mockYouTubeService.On("SearchVideoID", ctx, song).Return(videoID, nil)
		mockYouTubeService.On("GetVideoDetails", ctx, videoID).Return(youtubeMetadata, nil)
		mockAudioService.On("StartOperation", ctx, videoID).Return("test-operation-id", "test-video-id", nil)

		done := make(chan struct{})

		mockAudioService.On("ProcessAudio", mock.Anything, "test-operation-id", *youtubeMetadata).Return(nil).Run(func(args mock.Arguments) {
			go func() {
				defer close(done)
			}()
		})

		// Act
		operationID, songID, err := uc.Execute(ctx, song)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "test-operation-id", operationID)
		assert.Equal(t, "test-video-id", songID)
		<-done
		mockYouTubeService.AssertExpectations(t)
		mockAudioService.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockYouTubeService := new(MockYouTubeService)
		mockAudioService := new(MockAudioProcessingService)

		uc := usecase.NewInitiateDownloadUseCase(mockAudioService, mockYouTubeService)

		ctx := context.Background()
		song := "test song"
		videoID := "test-video-id"
		operationID := "test-operation-id"
		youtubeMetadata := &api.VideoDetails{
			VideoID:    "test-video-id",
			Title:      "Test Video",
			Duration:   "3:00",
			URLYouTube: "https://youtube.com/watch?v=test-video-id",
			Thumbnail:  "https://img.youtube.com/vi/test-video-id/0.jpg",
		}
		expectedError := errors.New("error procesando audio")

		mockYouTubeService.On("SearchVideoID", ctx, song).Return(videoID, nil)
		mockYouTubeService.On("GetVideoDetails", ctx, videoID).Return(youtubeMetadata, nil)
		mockAudioService.On("StartOperation", ctx, videoID).Return(operationID, "test-video-id", nil)

		done := make(chan struct{})

		mockAudioService.On("ProcessAudio", mock.Anything, operationID, *youtubeMetadata).Return(expectedError).Run(func(args mock.Arguments) {
			go func() {
				defer close(done)
			}()
		})

		// act
		operationIDResult, songID, err := uc.Execute(ctx, song)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, operationID, operationIDResult)
		assert.Equal(t, songID, "test-video-id")

		<-done

		mockYouTubeService.AssertExpectations(t)
		mockAudioService.AssertExpectations(t)
	})

	t.Run("Error in the search ID of the song", func(t *testing.T) {
		ctx := context.Background()
		song := "test song"

		mockYouTube := new(MockYouTubeService)
		mockAudioService := new(MockAudioProcessingService)

		uc := usecase.NewInitiateDownloadUseCase(mockAudioService, mockYouTube)

		mockYouTube.On("SearchVideoID", ctx, song).Return("", errors.New("error al buscar ID"))

		_, _, err := uc.Execute(ctx, song)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error al buscar el ID de la cancion")
	})

	t.Run("Error getting YouTube metadata", func(t *testing.T) {
		ctx := context.Background()
		song := "test song"
		mockYouTube := new(MockYouTubeService)
		mockAudioService := new(MockAudioProcessingService)

		uc := usecase.NewInitiateDownloadUseCase(mockAudioService, mockYouTube)

		videoID := "test-video-id"
		mockYouTube.On("SearchVideoID", ctx, song).Return(videoID, nil)
		mockYouTube.On("GetVideoDetails", ctx, videoID).Return(&api.VideoDetails{}, errors.New("error al obtener metadata"))

		_, _, err := uc.Execute(ctx, song)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error al obtener metadata de YouTube")
	})

	t.Run("Error starting operation", func(t *testing.T) {
		song := "test song"
		ctx := context.Background()

		mockYouTube := new(MockYouTubeService)
		mockAudioService := new(MockAudioProcessingService)

		uc := usecase.NewInitiateDownloadUseCase(mockAudioService, mockYouTube)

		videoID := "test-video-id"
		mockYouTube.On("SearchVideoID", ctx, song).Return(videoID, nil)
		mockYouTube.On("GetVideoDetails", ctx, videoID).Return(&api.VideoDetails{}, nil)
		mockAudioService.On("StartOperation", ctx, videoID).Return("", "", errors.New("error al iniciar operación"))

		_, _, err := uc.Execute(ctx, song)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error al iniciar la operación")
	})
}
