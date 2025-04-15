//go:build !integration

package service

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestMediaService_CreateMedia(t *testing.T) {
	mockRepo := new(MockMediaRepository)
	mockLogger := new(logger.MockLogger)

	service := NewMediaService(mockRepo, mockLogger)

	media := &model.Media{
		VideoID: "test-video-id",
		Status:  "starting",
		Metadata: &model.PlatformMetadata{
			Title: "test-title",
		},
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockRepo.On("SaveMedia", mock.Anything, media).Return(nil)

	// Act
	err := service.CreateMedia(context.Background(), media)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_CreateMedia_Error(t *testing.T) {
	// Arrange
	mockRepo := new(MockMediaRepository)
	mockLogger := new(logger.MockLogger)

	service := NewMediaService(mockRepo, mockLogger)

	media := &model.Media{
		VideoID: "test-video-id",
		Status:  "starting",
		Metadata: &model.PlatformMetadata{
			Title: "test-title",
		},
	}

	expectedError := errors.New("repository error")
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockRepo.On("SaveMedia", mock.Anything, media).Return(expectedError)

	// Act
	err := service.CreateMedia(context.Background(), media)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository error")
	mockRepo.AssertExpectations(t)
}

func TestMediaService_GetMediaByID(t *testing.T) {
	// Arrange
	mockRepo := new(MockMediaRepository)
	mockLogger := new(logger.MockLogger)

	service := NewMediaService(mockRepo, mockLogger)

	videoID := "test-video-id"
	expectedMedia := &model.Media{
		VideoID: videoID,
		Status:  "starting",
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockRepo.On("GetMediaByID", mock.Anything, videoID).Return(expectedMedia, nil)

	// Act
	media, err := service.GetMediaByID(context.Background(), videoID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expectedMedia, media)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_GetMediaByID_Error(t *testing.T) {
	// Arrange
	mockRepo := new(MockMediaRepository)
	mockLogger := new(logger.MockLogger)

	service := NewMediaService(mockRepo, mockLogger)

	videoID := "test-video-id"
	expectedError := errors.New("repository error")

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockRepo.On("GetMediaByID", mock.Anything, videoID).Return(&model.Media{}, expectedError)

	// Act
	media, err := service.GetMediaByID(context.Background(), videoID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Contains(t, err.Error(), "repository error")
	mockRepo.AssertExpectations(t)
}

func TestMediaService_UpdateMedia(t *testing.T) {
	// Arrange
	mockRepo := new(MockMediaRepository)
	mockLogger := new(logger.MockLogger)

	service := NewMediaService(mockRepo, mockLogger)

	videoID := "test-video-id"
	media := &model.Media{
		VideoID: videoID,
		Status:  "completed",
		Metadata: &model.PlatformMetadata{
			Title: "test-title",
		},
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockRepo.On("UpdateMedia", mock.Anything, videoID, media).Return(nil)

	// Act
	err := service.UpdateMedia(context.Background(), videoID, media)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_UpdateMedia_Error(t *testing.T) {
	// Arrange
	mockRepo := new(MockMediaRepository)
	mockLogger := new(logger.MockLogger)

	service := NewMediaService(mockRepo, mockLogger)

	videoID := "test-video-id"
	media := &model.Media{
		VideoID: videoID,
		Status:  "completed",
		Metadata: &model.PlatformMetadata{
			Title: "test-title",
		},
	}

	expectedError := errors.New("repository error")
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockRepo.On("UpdateMedia", mock.Anything, videoID, media).Return(expectedError)

	// Act
	err := service.UpdateMedia(context.Background(), videoID, media)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository error")
	mockRepo.AssertExpectations(t)
}

func TestNormalizedTitle(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Elimina barras", "test/title", "testtitle"},
		{"Elimina puntos", "test.title", "testtitle"},
		{"Elimina múltiples especiales", "test/title.with$special#chars", "testtitlewithspecialchars"},
		{"Mantiene espacios", "test title", "test title"},
		{"Convierte a minúsculas", "TEST Title", "test title"},
		{"Elimina emojis", "test 😊 title", "test  title"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizedTitle(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMediaService_DeleteMedia(t *testing.T) {
	// Arrange
	mockRepo := new(MockMediaRepository)
	mockLogger := new(logger.MockLogger)

	service := NewMediaService(mockRepo, mockLogger)

	videoID := "test-video-id"

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockRepo.On("DeleteMedia", mock.Anything, videoID).Return(nil)

	// Act
	err := service.DeleteMedia(context.Background(), videoID)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_DeleteMedia_Error(t *testing.T) {
	// Arrange
	mockRepo := new(MockMediaRepository)
	mockLogger := new(logger.MockLogger)

	service := NewMediaService(mockRepo, mockLogger)

	videoID := "test-video-id"
	expectedError := errors.New("repository error")

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockRepo.On("DeleteMedia", mock.Anything, videoID).Return(expectedError)

	// Act
	err := service.DeleteMedia(context.Background(), videoID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repository error")
	mockRepo.AssertExpectations(t)
}
