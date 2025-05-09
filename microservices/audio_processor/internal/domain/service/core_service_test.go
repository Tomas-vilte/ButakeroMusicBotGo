////go:build !integration

package service

import (
	"bytes"
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestCoreService_ProcessMedia_Success(t *testing.T) {
	mockMediaRepository := new(MockMediaRepository)
	mockAudioStorageService := new(MockAudioStorageService)
	mockTopicPublisher := new(MockMessageQueue)
	mockAudioDownloadService := new(MockAudioDownloadService)
	mockLogger := new(logger.MockLogger)

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Timeout:     30 * time.Second,
			MaxAttempts: 3,
		},
	}

	service := NewCoreService(
		mockMediaRepository,
		mockAudioStorageService,
		mockTopicPublisher,
		mockAudioDownloadService,
		mockLogger,
		cfg,
	)

	media := &model.Media{
		VideoID:    "test-video-id",
		TitleLower: "test song",
		Metadata: &model.PlatformMetadata{
			Title:        "Test Song",
			DurationMs:   123456,
			URL:          "https://example.com/test-song",
			ThumbnailURL: "https://example.com/test-thumbnail.jpg",
			Platform:     "youtube",
		},
	}

	userID := "user_123"
	interactionID := "interaction_123"

	audioBuffer := bytes.NewBuffer([]byte("test audio data"))
	fileData := &model.FileData{
		FilePath: "test-song.dca",
		FileSize: "1234",
		FileType: "audio/dca",
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockAudioDownloadService.On("DownloadAndEncode", mock.Anything, media.Metadata.URL).Return(audioBuffer, nil)
	mockAudioStorageService.On("StoreAudio", mock.Anything, audioBuffer, media.TitleLower).Return(fileData, nil)
	mockMediaRepository.On("UpdateMedia", mock.Anything, media.VideoID, mock.AnythingOfType("*model.Media")).Return(nil)
	mockTopicPublisher.On("Publish", mock.Anything, mock.AnythingOfType("*model.MediaProcessingMessage")).Return(nil)

	err := service.ProcessMedia(context.Background(), media, userID, interactionID)

	// Assert
	assert.NoError(t, err)
	mockAudioDownloadService.AssertExpectations(t)
	mockAudioStorageService.AssertExpectations(t)
	mockMediaRepository.AssertExpectations(t)
	mockTopicPublisher.AssertExpectations(t)
}

func TestCoreService_ProcessMedia_DownloadError(t *testing.T) {
	// Arrange
	mockMediaRepository := new(MockMediaRepository)
	mockAudioStorageService := new(MockAudioStorageService)
	mockTopicPublisher := new(MockMessageQueue)
	mockAudioDownloadService := new(MockAudioDownloadService)
	mockLogger := new(logger.MockLogger)

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Timeout:     30 * time.Second,
			MaxAttempts: 3,
		},
	}

	service := NewCoreService(
		mockMediaRepository,
		mockAudioStorageService,
		mockTopicPublisher,
		mockAudioDownloadService,
		mockLogger,
		cfg,
	)

	media := &model.Media{
		VideoID:    "test-video-id",
		TitleLower: "test song",
		Metadata: &model.PlatformMetadata{
			Title:        "Test Song",
			DurationMs:   123456,
			URL:          "https://example.com/test-song",
			ThumbnailURL: "https://example.com/test-thumbnail.jpg",
			Platform:     "youtube",
		},
	}

	userID := "user_123"
	interactionID := "interaction_123"

	expectedError := errors.New("download failed")
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockAudioDownloadService.On("DownloadAndEncode", mock.Anything, media.Metadata.URL).Return((*bytes.Buffer)(nil), expectedError)

	// Act
	err := service.ProcessMedia(context.Background(), media, userID, interactionID)

	// Assert
	assert.Error(t, err)

	assert.Equal(t, err.Error(), "número máximo de intentos alcanzado (3): download failed")

	mockAudioDownloadService.AssertExpectations(t)
	mockAudioStorageService.AssertNotCalled(t, "StoreAudio")
	mockMediaRepository.AssertNotCalled(t, "UpdateMedia")
	mockTopicPublisher.AssertNotCalled(t, "Publish")
}

func TestCoreService_ProcessMedia_StorageError(t *testing.T) {
	// Arrange
	mockMediaRepository := new(MockMediaRepository)
	mockAudioStorageService := new(MockAudioStorageService)
	mockTopicPublisher := new(MockMessageQueue)
	mockAudioDownloadService := new(MockAudioDownloadService)
	mockLogger := new(logger.MockLogger)

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Timeout:     30 * time.Second,
			MaxAttempts: 3,
		},
	}

	service := NewCoreService(
		mockMediaRepository,
		mockAudioStorageService,
		mockTopicPublisher,
		mockAudioDownloadService,
		mockLogger,
		cfg,
	)

	media := &model.Media{
		VideoID:    "test-video-id",
		TitleLower: "test song",
		Metadata: &model.PlatformMetadata{
			Title:        "Test Song",
			DurationMs:   123456,
			URL:          "https://example.com/test-song",
			ThumbnailURL: "https://example.com/test-thumbnail.jpg",
			Platform:     "youtube",
		},
	}

	userID := "user_123"
	interactionID := "interaction_123"

	audioBuffer := bytes.NewBuffer([]byte("test audio data"))
	expectedError := errors.New("storage failed")

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockAudioDownloadService.On("DownloadAndEncode", mock.Anything, media.Metadata.URL).Return(audioBuffer, nil)
	mockAudioStorageService.On("StoreAudio", mock.Anything, audioBuffer, media.TitleLower).Return((*model.FileData)(nil), expectedError)

	// Act
	err := service.ProcessMedia(context.Background(), media, userID, interactionID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "número máximo de intentos alcanzado (3)")
	mockAudioDownloadService.AssertExpectations(t)
	mockAudioStorageService.AssertExpectations(t)
	mockMediaRepository.AssertNotCalled(t, "UpdateMedia")
	mockTopicPublisher.AssertNotCalled(t, "Publish")
}

func TestCoreService_ProcessMedia_UpdateMediaError(t *testing.T) {
	// Arrange
	mockMediaRepository := new(MockMediaRepository)
	mockAudioStorageService := new(MockAudioStorageService)
	mockTopicPublisher := new(MockMessageQueue)
	mockAudioDownloadService := new(MockAudioDownloadService)
	mockLogger := new(logger.MockLogger)

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Timeout:     30 * time.Second,
			MaxAttempts: 3,
		},
	}

	service := NewCoreService(
		mockMediaRepository,
		mockAudioStorageService,
		mockTopicPublisher,
		mockAudioDownloadService,
		mockLogger,
		cfg,
	)

	media := &model.Media{
		VideoID:    "test-video-id",
		TitleLower: "test song",
		Metadata: &model.PlatformMetadata{
			Title:        "Test Song",
			DurationMs:   123456,
			URL:          "https://example.com/test-song",
			ThumbnailURL: "https://example.com/test-thumbnail.jpg",
			Platform:     "youtube",
		},
	}

	userID := "user_123"
	interactionID := "interaction_123"

	audioBuffer := bytes.NewBuffer([]byte("test audio data"))
	fileData := &model.FileData{
		FilePath: "test-song.dca",
		FileSize: "1234",
		FileType: "audio/dca",
	}
	expectedError := errors.New("update failed")

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockAudioDownloadService.On("DownloadAndEncode", mock.Anything, media.Metadata.URL).Return(audioBuffer, nil)
	mockAudioStorageService.On("StoreAudio", mock.Anything, audioBuffer, media.TitleLower).Return(fileData, nil)
	mockMediaRepository.On("UpdateMedia", mock.Anything, media.VideoID, mock.AnythingOfType("*model.Media")).Return(expectedError)

	// Act
	err := service.ProcessMedia(context.Background(), media, userID, interactionID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")

	mockAudioDownloadService.AssertExpectations(t)
	mockAudioStorageService.AssertExpectations(t)
	mockMediaRepository.AssertExpectations(t)
	mockTopicPublisher.AssertNotCalled(t, "Publish")
}

func TestCoreService_ProcessMedia_PublishError(t *testing.T) {
	// Arrange
	mockMediaRepository := new(MockMediaRepository)
	mockAudioStorageService := new(MockAudioStorageService)
	mockTopicPublisher := new(MockMessageQueue)
	mockAudioDownloadService := new(MockAudioDownloadService)
	mockLogger := new(logger.MockLogger)

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Timeout:     30 * time.Second,
			MaxAttempts: 3,
		},
	}

	service := NewCoreService(
		mockMediaRepository,
		mockAudioStorageService,
		mockTopicPublisher,
		mockAudioDownloadService,
		mockLogger,
		cfg,
	)

	media := &model.Media{
		VideoID:    "test-video-id",
		TitleLower: "test song",
		Metadata: &model.PlatformMetadata{
			Title:        "Test Song",
			DurationMs:   123456,
			URL:          "https://example.com/test-song",
			ThumbnailURL: "https://example.com/test-thumbnail.jpg",
			Platform:     "youtube",
		},
	}

	userID := "user_123"
	interactionID := "interaction_123"

	audioBuffer := bytes.NewBuffer([]byte("test audio data"))
	fileData := &model.FileData{
		FilePath: "test-song.dca",
		FileSize: "1234",
		FileType: "audio/dca",
	}
	expectedError := errors.New("publish failed")

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockAudioDownloadService.On("DownloadAndEncode", mock.Anything, media.Metadata.URL).Return(audioBuffer, nil)
	mockAudioStorageService.On("StoreAudio", mock.Anything, audioBuffer, media.TitleLower).Return(fileData, nil)
	mockMediaRepository.On("UpdateMedia", mock.Anything, media.VideoID, mock.AnythingOfType("*model.Media")).Return(nil)
	mockTopicPublisher.On("Publish", mock.Anything, mock.AnythingOfType("*model.MediaProcessingMessage")).Return(expectedError)

	// Act
	err := service.ProcessMedia(context.Background(), media, userID, interactionID)

	// Assert
	assert.Error(t, err)

	assert.Contains(t, err.Error(), "publish failed")

	mockAudioDownloadService.AssertExpectations(t)
	mockAudioStorageService.AssertExpectations(t)
	mockMediaRepository.AssertExpectations(t)
	mockTopicPublisher.AssertExpectations(t)
}
