package unit

import (
	"bytes"
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestAudioProcessingService(t *testing.T) {
	t.Run("StartOperation", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// arrange
			mockLogger := new(MockLogger)
			mockStorage := new(MockStorage)
			mockDownloader := new(MockDownloader)
			mockOperationRepo := new(MockOperationRepository)
			mockMetadataRepo := new(MockMetadataRepository)

			configService := config.Config{
				MaxAttempts: 3,
				Timeout:     5 * time.Minute,
			}

			serviceAudio := service.NewAudioProcessingService(mockLogger, mockStorage, mockDownloader, mockOperationRepo, mockMetadataRepo, configService)

			ctx := context.Background()
			operationID := "test-operation-id"
			youtubeMetadata := api.VideoDetails{
				VideoID:    "test-video-id",
				Title:      "Test Video",
				Duration:   "3:00",
				URLYouTube: "https://youtube.com/watch?v=test-video-id",
				Thumbnail:  "https://img.youtube.com/vi/test-video-id/0.jpg",
			}

			mockDownloader.On("DownloadAudio", mock.Anything, mock.AnythingOfType("string")).Return(bytes.NewReader([]byte("test audio data")), nil)
			mockStorage.On("UploadFile", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
			mockMetadataRepo.On("SaveMetadata", mock.Anything, mock.AnythingOfType("model.Metadata")).Return(nil)
			mockOperationRepo.On("SaveOperationsResult", mock.Anything, mock.AnythingOfType("model.OperationResult")).Return(nil)
			mockLogger.On("Info", mock.Anything, mock.Anything).Return()
			mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
			mockLogger.On("Error", mock.Anything, mock.Anything).Return()

			// act
			err := serviceAudio.ProcessAudio(ctx, operationID, youtubeMetadata)

			// assert
			assert.NoError(t, err)
			mockDownloader.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
			mockMetadataRepo.AssertExpectations(t)
			mockOperationRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)

		})
	})

	t.Run("FailureToSaveOperation", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockLogger)
		mockStorage := new(MockStorage)
		mockDownloader := new(MockDownloader)
		mockOperationRepo := new(MockOperationRepository)
		mockMetadataRepo := new(MockMetadataRepository)

		configService := config.Config{
			MaxAttempts: 3,
			Timeout:     5 * time.Minute,
		}

		serviceAudio := service.NewAudioProcessingService(mockLogger, mockStorage, mockDownloader, mockOperationRepo, mockMetadataRepo, configService)

		ctx := context.Background()
		songID := "test-song-id"

		mockOperationRepo.On("SaveOperationsResult", mock.Anything, mock.AnythingOfType("model.OperationResult")).Return(errors.New("database error"))

		// Act
		operationID, _, err := serviceAudio.StartOperation(ctx, songID)

		// Assert
		assert.Error(t, err)
		assert.Empty(t, operationID)
		assert.Contains(t, err.Error(), "error al guardar operación")
		mockOperationRepo.AssertExpectations(t)
	})

	t.Run("ProcessAudio", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			// arrange
			mockLogger := new(MockLogger)
			mockStorage := new(MockStorage)
			mockDownloader := new(MockDownloader)
			mockOperationRepo := new(MockOperationRepository)
			mockMetadataRepo := new(MockMetadataRepository)

			configService := config.Config{
				MaxAttempts: 3,
				Timeout:     5 * time.Minute,
			}

			serviceAudio := service.NewAudioProcessingService(mockLogger, mockStorage, mockDownloader, mockOperationRepo, mockMetadataRepo, configService)

			ctx := context.Background()
			operationID := "test-operation-id"
			youtubeMetadata := api.VideoDetails{
				VideoID:    "test-video-id",
				Title:      "Test Video",
				Duration:   "3:00",
				URLYouTube: "https://youtube.com/watch?v=test-video-id",
				Thumbnail:  "https://img.youtube.com/vi/test-video-id/0.jpg",
			}

			mockDownloader.On("DownloadAudio", mock.Anything, mock.AnythingOfType("string")).Return(bytes.NewReader([]byte("test audio data")), nil)
			mockStorage.On("UploadFile", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
			mockMetadataRepo.On("SaveMetadata", mock.Anything, mock.AnythingOfType("model.Metadata")).Return(nil)
			mockOperationRepo.On("SaveOperationsResult", mock.Anything, mock.AnythingOfType("model.OperationResult")).Return(nil)
			mockLogger.On("Info", mock.Anything, mock.Anything).Return()
			mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
			mockLogger.On("Error", mock.Anything, mock.Anything).Return()

			// Act
			err := serviceAudio.ProcessAudio(ctx, operationID, youtubeMetadata)

			// Assert
			assert.NoError(t, err)
			mockDownloader.AssertExpectations(t)
			mockStorage.AssertExpectations(t)
			mockMetadataRepo.AssertExpectations(t)
			mockOperationRepo.AssertExpectations(t)
			mockLogger.AssertExpectations(t)
		})
	})

	t.Run("DownloadError", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockLogger)
		mockStorage := new(MockStorage)
		mockDownloader := new(MockDownloader)
		mockOperationRepo := new(MockOperationRepository)
		mockMetadataRepo := new(MockMetadataRepository)

		configService := config.Config{
			MaxAttempts: 3,
			Timeout:     5 * time.Minute,
		}

		serviceAudio := service.NewAudioProcessingService(mockLogger, mockStorage, mockDownloader, mockOperationRepo, mockMetadataRepo, configService)

		ctx := context.Background()
		operationID := "test-operation-id"
		youtubeMetadata := api.VideoDetails{
			VideoID:    "test-video-id",
			Title:      "Test Video",
			Duration:   "3:00",
			URLYouTube: "https://youtube.com/watch?v=test-video-id",
			Thumbnail:  "https://img.youtube.com/vi/test-video-id/0.jpg",
		}
		mockAudioContent := bytes.NewBufferString("fake audio content")

		mockDownloader.On("DownloadAudio", mock.Anything, mock.AnythingOfType("string")).Return(mockAudioContent, errors.New("download error"))
		mockOperationRepo.On("SaveOperationsResult", mock.Anything, mock.AnythingOfType("model.OperationResult")).Return(nil)
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()

		// Act
		err := serviceAudio.ProcessAudio(ctx, operationID, youtubeMetadata)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "el procesamiento falló después de 3 intentos")
		mockDownloader.AssertExpectations(t)
		mockOperationRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("UploadError", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockLogger)
		mockStorage := new(MockStorage)
		mockDownloader := new(MockDownloader)
		mockOperationRepo := new(MockOperationRepository)
		mockMetadataRepo := new(MockMetadataRepository)

		configService := config.Config{
			MaxAttempts: 3,
			Timeout:     5 * time.Minute,
		}

		serviceAudio := service.NewAudioProcessingService(mockLogger, mockStorage, mockDownloader, mockOperationRepo, mockMetadataRepo, configService)

		ctx := context.Background()
		operationID := "test-operation-id"
		youtubeMetadata := api.VideoDetails{
			VideoID:    "test-video-id",
			Title:      "Test Video",
			Duration:   "3:00",
			URLYouTube: "https://youtube.com/watch?v=test-video-id",
			Thumbnail:  "https://img.youtube.com/vi/test-video-id/0.jpg",
		}

		mockDownloader.On("DownloadAudio", mock.Anything, mock.AnythingOfType("string")).Return(bytes.NewReader([]byte("test audio data")), nil)
		mockStorage.On("UploadFile", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
		mockMetadataRepo.On("SaveMetadata", mock.Anything, mock.AnythingOfType("model.Metadata")).Return(errors.New("metadata save error"))
		mockOperationRepo.On("SaveOperationsResult", mock.Anything, mock.AnythingOfType("model.OperationResult")).Return(nil)
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

		// Act
		err := serviceAudio.ProcessAudio(ctx, operationID, youtubeMetadata)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "el procesamiento falló después de 3 intentos")
		mockDownloader.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
		mockMetadataRepo.AssertExpectations(t)
		mockOperationRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("SaveMetadataError", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockLogger)
		mockStorage := new(MockStorage)
		mockDownloader := new(MockDownloader)
		mockOperationRepo := new(MockOperationRepository)
		mockMetadataRepo := new(MockMetadataRepository)

		configService := config.Config{
			MaxAttempts: 3,
			Timeout:     5 * time.Minute,
		}

		serviceAudio := service.NewAudioProcessingService(mockLogger, mockStorage, mockDownloader, mockOperationRepo, mockMetadataRepo, configService)

		ctx := context.Background()
		operationID := "test-operation-id"
		youtubeMetadata := api.VideoDetails{
			VideoID:    "test-video-id",
			Title:      "Test Video",
			Duration:   "3:00",
			URLYouTube: "https://youtube.com/watch?v=test-video-id",
			Thumbnail:  "https://img.youtube.com/vi/test-video-id/0.jpg",
		}

		mockDownloader.On("DownloadAudio", mock.Anything, mock.AnythingOfType("string")).Return(bytes.NewReader([]byte("test audio data")), nil)
		mockStorage.On("UploadFile", mock.Anything, mock.AnythingOfType("string"), mock.Anything).Return(nil)
		mockMetadataRepo.On("SaveMetadata", mock.Anything, mock.AnythingOfType("model.Metadata")).Return(errors.New("metadata save error"))
		mockOperationRepo.On("SaveOperationsResult", mock.Anything, mock.AnythingOfType("model.OperationResult")).Return(nil)
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

		// Act
		err := serviceAudio.ProcessAudio(ctx, operationID, youtubeMetadata)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "el procesamiento falló después de 3 intentos")
		mockDownloader.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
		mockMetadataRepo.AssertExpectations(t)
		mockOperationRepo.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
