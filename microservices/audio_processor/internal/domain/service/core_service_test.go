//go:build !integration

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
	mockMediaService := new(MockMediaService)
	mockAudioStorageService := new(MockAudioStorageService)
	mockTopicPublisher := new(MockTopicPublisherService)
	mockAudioDownloadService := new(MockAudioDownloadService)
	mockLogger := new(logger.MockLogger)

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Timeout:     30 * time.Second,
			MaxAttempts: 3,
		},
	}

	service := NewCoreService(
		mockMediaService,
		mockAudioStorageService,
		mockTopicPublisher,
		mockAudioDownloadService,
		mockLogger,
		cfg,
	)

	mediaDetails := &model.MediaDetails{
		ID:           "test-video-id",
		Title:        "Test Song",
		DurationMs:   123456,
		URL:          "https://example.com/test-song",
		ThumbnailURL: "https://example.com/test-thumbnail.jpg",
		Provider:     "youtube",
	}

	audioBuffer := bytes.NewBuffer([]byte("test audio data"))
	fileData := &model.FileData{
		FilePath: "test-song.dca",
		FileSize: "1234",
		FileType: "audio/dca",
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockAudioDownloadService.On("DownloadAndEncode", mock.Anything, mediaDetails.URL).Return(audioBuffer, nil)
	mockAudioStorageService.On("StoreAudio", mock.Anything, audioBuffer, mediaDetails.Title).Return(fileData, nil)
	mockMediaService.On("UpdateMedia", mock.Anything, mediaDetails.ID, mock.AnythingOfType("*model.Media")).Return(nil)
	mockTopicPublisher.On("PublishMediaProcessed", mock.Anything, mock.AnythingOfType("*model.MediaProcessingMessage")).Return(nil)

	err := service.ProcessMedia(context.Background(), mediaDetails)

	// Assert
	assert.NoError(t, err)
	mockAudioDownloadService.AssertExpectations(t)
	mockAudioStorageService.AssertExpectations(t)
	mockMediaService.AssertExpectations(t)
	mockTopicPublisher.AssertExpectations(t)
}

func TestCoreService_ProcessMedia_DownloadError(t *testing.T) {
	// Arrange
	mockMediaService := new(MockMediaService)
	mockAudioStorageService := new(MockAudioStorageService)
	mockTopicPublisher := new(MockTopicPublisherService)
	mockAudioDownloadService := new(MockAudioDownloadService)
	mockLogger := new(logger.MockLogger)

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Timeout:     30 * time.Second,
			MaxAttempts: 3,
		},
	}

	service := NewCoreService(
		mockMediaService,
		mockAudioStorageService,
		mockTopicPublisher,
		mockAudioDownloadService,
		mockLogger,
		cfg,
	)

	mediaDetails := &model.MediaDetails{
		ID:           "test-video-id",
		Title:        "Test Song",
		DurationMs:   123456,
		URL:          "https://example.com/test-song",
		ThumbnailURL: "https://example.com/test-thumbnail.jpg",
		Provider:     "youtube",
	}

	expectedError := errors.New("download failed")
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockAudioDownloadService.On("DownloadAndEncode", mock.Anything, mediaDetails.URL).Return((*bytes.Buffer)(nil), expectedError)

	// Act
	err := service.ProcessMedia(context.Background(), mediaDetails)

	// Assert
	assert.Error(t, err)

	assert.Equal(t, err.Error(), "número máximo de intentos alcanzado (3): download failed")

	mockAudioDownloadService.AssertExpectations(t)
	mockAudioStorageService.AssertNotCalled(t, "StoreAudio")
	mockMediaService.AssertNotCalled(t, "UpdateMedia")
	mockTopicPublisher.AssertNotCalled(t, "PublishMediaProcessed")
}

func TestCoreService_ProcessMedia_StorageError(t *testing.T) {
	// Arrange
	mockMediaService := new(MockMediaService)
	mockAudioStorageService := new(MockAudioStorageService)
	mockTopicPublisher := new(MockTopicPublisherService)
	mockAudioDownloadService := new(MockAudioDownloadService)
	mockLogger := new(logger.MockLogger)

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Timeout:     30 * time.Second,
			MaxAttempts: 3,
		},
	}

	service := NewCoreService(
		mockMediaService,
		mockAudioStorageService,
		mockTopicPublisher,
		mockAudioDownloadService,
		mockLogger,
		cfg,
	)

	mediaDetails := &model.MediaDetails{
		ID:           "test-video-id",
		Title:        "Test Song",
		DurationMs:   123456,
		URL:          "https://example.com/test-song",
		ThumbnailURL: "https://example.com/test-thumbnail.jpg",
		Provider:     "youtube",
	}

	audioBuffer := bytes.NewBuffer([]byte("test audio data"))
	expectedError := errors.New("storage failed")

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockAudioDownloadService.On("DownloadAndEncode", mock.Anything, mediaDetails.URL).Return(audioBuffer, nil)
	mockAudioStorageService.On("StoreAudio", mock.Anything, audioBuffer, mediaDetails.Title).Return((*model.FileData)(nil), expectedError)

	// Act
	err := service.ProcessMedia(context.Background(), mediaDetails)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "número máximo de intentos alcanzado (3)")
	mockAudioDownloadService.AssertExpectations(t)
	mockAudioStorageService.AssertExpectations(t)
	mockMediaService.AssertNotCalled(t, "UpdateMedia")
	mockTopicPublisher.AssertNotCalled(t, "PublishMediaProcessed")
}

func TestCoreService_ProcessMedia_UpdateMediaError(t *testing.T) {
	// Arrange
	mockMediaService := new(MockMediaService)
	mockAudioStorageService := new(MockAudioStorageService)
	mockTopicPublisher := new(MockTopicPublisherService)
	mockAudioDownloadService := new(MockAudioDownloadService)
	mockLogger := new(logger.MockLogger)

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Timeout:     30 * time.Second,
			MaxAttempts: 3,
		},
	}

	service := NewCoreService(
		mockMediaService,
		mockAudioStorageService,
		mockTopicPublisher,
		mockAudioDownloadService,
		mockLogger,
		cfg,
	)

	mediaDetails := &model.MediaDetails{
		ID:           "test-video-id",
		Title:        "Test Song",
		DurationMs:   123456,
		URL:          "https://example.com/test-song",
		ThumbnailURL: "https://example.com/test-thumbnail.jpg",
		Provider:     "youtube",
	}

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
	mockAudioDownloadService.On("DownloadAndEncode", mock.Anything, mediaDetails.URL).Return(audioBuffer, nil)
	mockAudioStorageService.On("StoreAudio", mock.Anything, audioBuffer, mediaDetails.Title).Return(fileData, nil)
	mockMediaService.On("UpdateMedia", mock.Anything, mediaDetails.ID, mock.AnythingOfType("*model.Media")).Return(expectedError)

	// Act
	err := service.ProcessMedia(context.Background(), mediaDetails)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update failed")

	mockAudioDownloadService.AssertExpectations(t)
	mockAudioStorageService.AssertExpectations(t)
	mockMediaService.AssertExpectations(t)
	mockTopicPublisher.AssertNotCalled(t, "PublishMediaProcessed")
}

func TestCoreService_ProcessMedia_PublishError(t *testing.T) {
	// Arrange
	mockMediaService := new(MockMediaService)
	mockAudioStorageService := new(MockAudioStorageService)
	mockTopicPublisher := new(MockTopicPublisherService)
	mockAudioDownloadService := new(MockAudioDownloadService)
	mockLogger := new(logger.MockLogger)

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Timeout:     30 * time.Second,
			MaxAttempts: 3,
		},
	}

	service := NewCoreService(
		mockMediaService,
		mockAudioStorageService,
		mockTopicPublisher,
		mockAudioDownloadService,
		mockLogger,
		cfg,
	)

	mediaDetails := &model.MediaDetails{
		ID:           "test-video-id",
		Title:        "Test Song",
		DurationMs:   123456,
		URL:          "https://example.com/test-song",
		ThumbnailURL: "https://example.com/test-thumbnail.jpg",
		Provider:     "youtube",
	}

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
	mockAudioDownloadService.On("DownloadAndEncode", mock.Anything, mediaDetails.URL).Return(audioBuffer, nil)
	mockAudioStorageService.On("StoreAudio", mock.Anything, audioBuffer, mediaDetails.Title).Return(fileData, nil)
	mockMediaService.On("UpdateMedia", mock.Anything, mediaDetails.ID, mock.AnythingOfType("*model.Media")).Return(nil)
	mockTopicPublisher.On("PublishMediaProcessed", mock.Anything, mock.AnythingOfType("*model.MediaProcessingMessage")).Return(expectedError)

	// Act
	err := service.ProcessMedia(context.Background(), mediaDetails)

	// Assert
	assert.Error(t, err)

	assert.Contains(t, err.Error(), "publish failed")

	mockAudioDownloadService.AssertExpectations(t)
	mockAudioStorageService.AssertExpectations(t)
	mockMediaService.AssertExpectations(t)
	mockTopicPublisher.AssertExpectations(t)
}
