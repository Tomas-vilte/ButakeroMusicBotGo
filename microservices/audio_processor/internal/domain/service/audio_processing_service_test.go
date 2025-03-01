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
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAudioProcessingService_ProcessAudio_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	operationID := "test-operation-123"
	mediaDetails := &model.MediaDetails{
		ID:        "video123",
		Title:     "Test Video",
		Duration:  "3:45",
		URL:       "https://youtube.com/watch?v=video123",
		Thumbnail: "https://youtube.com/thumbnail.jpg",
	}

	mockDownloadService := new(MockAudioDownloadService)
	mockStorageService := new(MockAudioStorageService)
	mockOpsManager := new(MockOperationsManager)
	mockMessagingManager := new(MockMessagingManager)
	mockErrorHandler := new(MockErrorManagement)
	mockLogger := new(logger.MockLogger)

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Timeout:     30 * time.Second,
			MaxAttempts: 1,
		},
	}

	// Respuestas esperadas de los mocks
	audioBuffer := bytes.NewBuffer([]byte("test audio data"))
	fileData := &model.FileData{
		FilePath: "s3://bucket/audio/video123.dca",
		FileSize: "1024",
		FileType: ".dca",
	}

	// Configuración de expectativas de los mocks
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockDownloadService.On("DownloadAndEncode", mock.Anything, mediaDetails.URL).Return(audioBuffer, nil)
	mockStorageService.On("StoreAudio", mock.Anything, audioBuffer, mock.AnythingOfType("*model.Metadata")).Return(fileData, nil)
	mockOpsManager.On("HandleOperationSuccess", mock.Anything, operationID, mock.AnythingOfType("*model.Metadata"), fileData).Return(nil)
	mockMessagingManager.On("SendProcessingMessage", mock.Anything, operationID, statusSuccess, mock.AnythingOfType("*model.Metadata"), 1).Return(nil)

	service := NewAudioProcessingService(
		mockDownloadService,
		mockStorageService,
		mockOpsManager,
		mockMessagingManager,
		mockErrorHandler,
		mockLogger,
		cfg,
	)

	// Act
	err := service.ProcessAudio(ctx, operationID, mediaDetails)

	// Assert
	require.NoError(t, err)
	mockDownloadService.AssertExpectations(t)
	mockStorageService.AssertExpectations(t)
	mockOpsManager.AssertExpectations(t)
	mockMessagingManager.AssertExpectations(t)
	mockErrorHandler.AssertNotCalled(t, "HandleProcessingError")
}

func TestAudioProcessingService_ProcessAudio_DownloadError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	operationID := "test-operation-123"
	mediaDetails := &model.MediaDetails{
		ID:        "video123",
		Title:     "Test Video",
		Duration:  "3:45",
		URL:       "https://youtube.com/watch?v=video123",
		Thumbnail: "https://youtube.com/thumbnail.jpg",
	}

	mockDownloadService := new(MockAudioDownloadService)
	mockStorageService := new(MockAudioStorageService)
	mockOpsManager := new(MockOperationsManager)
	mockMessagingManager := new(MockMessagingManager)
	mockErrorHandler := new(MockErrorManagement)
	mockLogger := new(logger.MockLogger)

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Timeout:     100 * time.Millisecond,
			MaxAttempts: 1,
		},
	}

	expectedError := errors.New("download failed")
	audioBuffer := bytes.NewBuffer([]byte{})

	// Error en la descarga
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockDownloadService.On("DownloadAndEncode", mock.Anything, mediaDetails.URL).Return(audioBuffer, expectedError)
	mockErrorHandler.On("HandleProcessingError", mock.Anything, operationID, mock.AnythingOfType("*model.Metadata"), "download/encode", 1, expectedError).Return(expectedError)

	service := NewAudioProcessingService(
		mockDownloadService,
		mockStorageService,
		mockOpsManager,
		mockMessagingManager,
		mockErrorHandler,
		mockLogger,
		cfg,
	)

	// Act
	err := service.ProcessAudio(ctx, operationID, mediaDetails)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockDownloadService.AssertExpectations(t)
	mockStorageService.AssertNotCalled(t, "StoreAudio")
	mockOpsManager.AssertNotCalled(t, "HandleOperationSuccess")
	mockMessagingManager.AssertNotCalled(t, "SendProcessingMessage")
	mockErrorHandler.AssertExpectations(t)
}

func TestAudioProcessingService_ProcessAudio_StorageError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	operationID := "test-operation-123"
	mediaDetails := &model.MediaDetails{
		ID:        "video123",
		Title:     "Test Video",
		Duration:  "3:45",
		URL:       "https://youtube.com/watch?v=video123",
		Thumbnail: "https://youtube.com/thumbnail.jpg",
	}

	mockDownloadService := new(MockAudioDownloadService)
	mockStorageService := new(MockAudioStorageService)
	mockOpsManager := new(MockOperationsManager)
	mockMessagingManager := new(MockMessagingManager)
	mockErrorHandler := new(MockErrorManagement)
	mockLogger := new(logger.MockLogger)

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Timeout:     100 * time.Millisecond,
			MaxAttempts: 1,
		},
	}

	expectedError := errors.New("storage failed")
	audioBuffer := bytes.NewBuffer([]byte("test audio data"))
	fileData := &model.FileData{}

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockDownloadService.On("DownloadAndEncode", mock.Anything, mediaDetails.URL).Return(audioBuffer, nil)
	mockStorageService.On("StoreAudio", mock.Anything, audioBuffer, mock.AnythingOfType("*model.Metadata")).Return(fileData, expectedError)
	mockErrorHandler.On("HandleProcessingError", mock.Anything, operationID, mock.AnythingOfType("*model.Metadata"), "storage", 1, expectedError).Return(expectedError)

	service := NewAudioProcessingService(
		mockDownloadService,
		mockStorageService,
		mockOpsManager,
		mockMessagingManager,
		mockErrorHandler,
		mockLogger,
		cfg,
	)

	// Act
	err := service.ProcessAudio(ctx, operationID, mediaDetails)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockDownloadService.AssertExpectations(t)
	mockStorageService.AssertExpectations(t)
	mockOpsManager.AssertNotCalled(t, "HandleOperationSuccess")
	mockMessagingManager.AssertNotCalled(t, "SendProcessingMessage")
	mockErrorHandler.AssertExpectations(t)
}

func TestAudioProcessingService_ProcessAudio_OperationUpdateError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	operationID := "test-operation-123"
	mediaDetails := &model.MediaDetails{
		ID:        "video123",
		Title:     "Test Video",
		Duration:  "3:45",
		URL:       "https://youtube.com/watch?v=video123",
		Thumbnail: "https://youtube.com/thumbnail.jpg",
	}

	mockDownloadService := new(MockAudioDownloadService)
	mockStorageService := new(MockAudioStorageService)
	mockOpsManager := new(MockOperationsManager)
	mockMessagingManager := new(MockMessagingManager)
	mockErrorHandler := new(MockErrorManagement)
	mockLogger := new(logger.MockLogger)

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Timeout:     100 * time.Millisecond,
			MaxAttempts: 1,
		},
	}

	expectedError := errors.New("operation update failed")
	audioBuffer := bytes.NewBuffer([]byte("test audio data"))
	fileData := &model.FileData{
		FilePath: "s3://bucket/audio/video123.dca",
		FileSize: "1024",
		FileType: ".dca",
	}

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockDownloadService.On("DownloadAndEncode", mock.Anything, mediaDetails.URL).Return(audioBuffer, nil)
	mockStorageService.On("StoreAudio", mock.Anything, audioBuffer, mock.AnythingOfType("*model.Metadata")).Return(fileData, nil)
	mockOpsManager.On("HandleOperationSuccess", mock.Anything, operationID, mock.AnythingOfType("*model.Metadata"), fileData).Return(expectedError)
	mockErrorHandler.On("HandleProcessingError", mock.Anything, operationID, mock.AnythingOfType("*model.Metadata"), "operation_update", 1, expectedError).Return(expectedError)

	service := NewAudioProcessingService(
		mockDownloadService,
		mockStorageService,
		mockOpsManager,
		mockMessagingManager,
		mockErrorHandler,
		mockLogger,
		cfg,
	)

	// Act
	err := service.ProcessAudio(ctx, operationID, mediaDetails)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockDownloadService.AssertExpectations(t)
	mockStorageService.AssertExpectations(t)
	mockOpsManager.AssertExpectations(t)
	mockMessagingManager.AssertNotCalled(t, "SendProcessingMessage")
	mockErrorHandler.AssertExpectations(t)
}

func TestAudioProcessingService_ProcessAudio_Retry(t *testing.T) {
	// Arrange
	ctx := context.Background()
	operationID := "test-operation-123"
	mediaDetails := &model.MediaDetails{
		ID:        "video123",
		Title:     "Test Video",
		Duration:  "3:45",
		URL:       "https://youtube.com/watch?v=video123",
		Thumbnail: "https://youtube.com/thumbnail.jpg",
	}

	mockDownloadService := new(MockAudioDownloadService)
	mockStorageService := new(MockAudioStorageService)
	mockOpsManager := new(MockOperationsManager)
	mockMessagingManager := new(MockMessagingManager)
	mockErrorHandler := new(MockErrorManagement)
	mockLogger := new(logger.MockLogger)

	cfg := &config.Config{
		Service: config.ServiceConfig{
			Timeout:     5 * time.Second,
			MaxAttempts: 2,
		},
	}

	tempError := errors.New("temporary error")
	audioBuffer := bytes.NewBuffer([]byte("test audio data"))
	fileData := &model.FileData{
		FilePath: "s3://bucket/audio/video123.dca",
		FileSize: "1024",
		FileType: ".dca",
	}

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockDownloadService.On("DownloadAndEncode", mock.Anything, mediaDetails.URL).
		Return((*bytes.Buffer)(nil), tempError).Once()
	mockErrorHandler.On("HandleProcessingError",
		mock.Anything,
		operationID,
		mock.AnythingOfType("*model.Metadata"),
		"download/encode",
		1,
		tempError).
		Return(tempError).Once()

	// Segundo intento exitoso
	mockDownloadService.On("DownloadAndEncode", mock.Anything, mediaDetails.URL).
		Return(audioBuffer, nil).Once()
	mockStorageService.On("StoreAudio",
		mock.Anything,
		audioBuffer,
		mock.AnythingOfType("*model.Metadata")).
		Return(fileData, nil).Once()
	mockOpsManager.On("HandleOperationSuccess",
		mock.Anything,
		operationID,
		mock.AnythingOfType("*model.Metadata"),
		fileData).
		Return(nil).Once()
	mockMessagingManager.On("SendProcessingMessage",
		mock.Anything,
		operationID,
		statusSuccess,
		mock.AnythingOfType("*model.Metadata"),
		2).
		Return(nil).Once()

	service := NewAudioProcessingService(
		mockDownloadService,
		mockStorageService,
		mockOpsManager,
		mockMessagingManager,
		mockErrorHandler,
		mockLogger,
		cfg,
	)

	// Act
	err := service.ProcessAudio(ctx, operationID, mediaDetails)

	// Assert
	require.NoError(t, err)
	mockDownloadService.AssertExpectations(t)
	mockStorageService.AssertExpectations(t)
	mockOpsManager.AssertExpectations(t)
	mockMessagingManager.AssertExpectations(t)
	mockErrorHandler.AssertExpectations(t)
}

func TestAudioProcessingService_CreateMetadata(t *testing.T) {
	// Arrange
	mediaDetails := &model.MediaDetails{
		ID:        "video123",
		Title:     "Test Video",
		Duration:  "3:45",
		URL:       "https://youtube.com/watch?v=video123",
		Thumbnail: "https://youtube.com/thumbnail.jpg",
	}

	service := NewAudioProcessingService(
		nil, nil, nil, nil, nil, &logger.MockLogger{}, &config.Config{},
	)

	// Act
	metadata := service.createMetadata(mediaDetails)

	// Assert
	assert.NotEmpty(t, metadata.ID)
	assert.Equal(t, mediaDetails.ID, metadata.VideoID)
	assert.Equal(t, mediaDetails.Title, metadata.Title)
	assert.Equal(t, mediaDetails.Duration, metadata.Duration)
	assert.Equal(t, mediaDetails.URL, metadata.URL)
	assert.Equal(t, platformYoutube, metadata.Platform)
	assert.Equal(t, mediaDetails.Thumbnail, metadata.ThumbnailURL)
}
