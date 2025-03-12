package service

import (
	"context"
	"errors"
	"fmt"
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
		ID:      "test-id",
		VideoID: "test-video-id",
		Status:  "starting",
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
		ID:      "test-id",
		VideoID: "test-video-id",
		Status:  "starting",
	}

	expectedError := errors.New("repository error")
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockRepo.On("SaveMedia", mock.Anything, media).Return(expectedError)

	// Act
	err := service.CreateMedia(context.Background(), media)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("error al crear el registro de media: %w", expectedError), err)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_GetMediaByID(t *testing.T) {
	// Arrange
	mockRepo := new(MockMediaRepository)
	mockLogger := new(logger.MockLogger)

	service := NewMediaService(mockRepo, mockLogger)

	id := "test-id"
	videoID := "test-video-id"
	expectedMedia := &model.Media{
		ID:      id,
		VideoID: videoID,
		Status:  "starting",
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockRepo.On("GetMedia", mock.Anything, id, videoID).Return(expectedMedia, nil)

	// Act
	media, err := service.GetMediaByID(context.Background(), id, videoID)

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

	id := "test-id"
	videoID := "test-video-id"
	expectedError := errors.New("repository error")

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockRepo.On("GetMedia", mock.Anything, id, videoID).Return(&model.Media{}, expectedError)

	// Act
	media, err := service.GetMediaByID(context.Background(), id, videoID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, media)
	assert.Equal(t, fmt.Errorf("error al obtener el registro de media: %w", expectedError), err)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_UpdateMedia(t *testing.T) {
	// Arrange
	mockRepo := new(MockMediaRepository)
	mockLogger := new(logger.MockLogger)

	service := NewMediaService(mockRepo, mockLogger)

	id := "test-id"
	videoID := "test-video-id"
	media := &model.Media{
		ID:      id,
		VideoID: videoID,
		Status:  "completed",
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockRepo.On("UpdateMedia", mock.Anything, id, videoID, media).Return(nil)

	// Act
	err := service.UpdateMedia(context.Background(), id, videoID, media)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_UpdateMedia_Error(t *testing.T) {
	// Arrange
	mockRepo := new(MockMediaRepository)
	mockLogger := new(logger.MockLogger)

	service := NewMediaService(mockRepo, mockLogger)

	id := "test-id"
	videoID := "test-video-id"
	media := &model.Media{
		ID:      id,
		VideoID: videoID,
		Status:  "completed",
	}

	expectedError := errors.New("repository error")
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockRepo.On("UpdateMedia", mock.Anything, id, videoID, media).Return(expectedError)

	// Act
	err := service.UpdateMedia(context.Background(), id, videoID, media)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("error al actualizar el estado del registro de media: %w", expectedError), err)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_DeleteMedia(t *testing.T) {
	// Arrange
	mockRepo := new(MockMediaRepository)
	mockLogger := new(logger.MockLogger)

	service := NewMediaService(mockRepo, mockLogger)

	id := "test-id"
	videoID := "test-video-id"

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockRepo.On("DeleteMedia", mock.Anything, id, videoID).Return(nil)

	// Act
	err := service.DeleteMedia(context.Background(), id, videoID)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestMediaService_DeleteMedia_Error(t *testing.T) {
	// Arrange
	mockRepo := new(MockMediaRepository)
	mockLogger := new(logger.MockLogger)

	service := NewMediaService(mockRepo, mockLogger)

	id := "test-id"
	videoID := "test-video-id"
	expectedError := errors.New("repository error")

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockRepo.On("DeleteMedia", mock.Anything, id, videoID).Return(expectedError)

	// Act
	err := service.DeleteMedia(context.Background(), id, videoID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, fmt.Errorf("error al eliminar el registro de media: %w", expectedError), err)
	mockRepo.AssertExpectations(t)
}
