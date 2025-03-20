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
	"time"
)

func TestOperationService_StartOperation(t *testing.T) {
	// Arrange
	mockRepo := new(MockMediaRepository)
	mockLogger := new(logger.MockLogger)

	service := NewOperationService(mockRepo, mockLogger)

	videoID := "test-video-id"
	expectedMedia := &model.Media{
		VideoID:    videoID,
		Status:     "starting",
		TitleLower: "",
		Metadata: &model.PlatformMetadata{
			DurationMs:   0,
			URL:          "",
			ThumbnailURL: "",
			Platform:     "",
		},
		FileData: &model.FileData{
			FilePath: "",
			FileSize: "",
			FileType: "",
		},
		ProcessingDate: time.Now(),
		Success:        false,
		Attempts:       0,
		Failures:       0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		PlayCount:      0,
	}

	mockRepo.On("SaveMedia", mock.Anything, mock.MatchedBy(func(media *model.Media) bool {
		return media.VideoID == videoID && media.Status == "starting"
	})).Return(nil)

	// Act
	result, err := service.StartOperation(context.Background(), videoID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedMedia.VideoID, result.VideoID)
	assert.Equal(t, expectedMedia.Status, result.Status)

	mockRepo.AssertExpectations(t)
}

func TestOperationService_StartOperation_Error(t *testing.T) {
	// Arrange
	mockRepo := new(MockMediaRepository)
	mockLogger := new(logger.MockLogger)

	service := NewOperationService(mockRepo, mockLogger)

	videoID := "test-video-id"
	expectedError := errors.New("repository error")

	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockRepo.On("SaveMedia", mock.Anything, mock.AnythingOfType("*model.Media")).Return(expectedError)

	// Act
	result, err := service.StartOperation(context.Background(), videoID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "repository error")

	mockRepo.AssertExpectations(t)
}
