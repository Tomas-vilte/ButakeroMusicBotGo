//go:build !integration

package processor

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestNewMediaProcessor(t *testing.T) {
	mockMediaRepo := new(MockMediaRepository)
	mockVideoService := new(MockVideoService)
	mockCoreService := new(MockCoreService)
	mockLogger := new(logger.MockLogger)

	processor := NewMediaProcessor(
		mockMediaRepo,
		mockVideoService,
		mockCoreService,
		mockLogger,
	)

	assert.NotNil(t, processor)
	assert.Equal(t, mockMediaRepo, processor.mediaRepo)
	assert.Equal(t, mockVideoService, processor.videoService)
	assert.Equal(t, mockCoreService, processor.coreService)
	assert.Equal(t, mockLogger, processor.logger)
}

func TestMediaProcessor_ProcessRequest_Success(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockVideoService := new(MockVideoService)
	mockCoreService := new(MockCoreService)
	mockLogger := new(logger.MockLogger)

	processor := NewMediaProcessor(
		mockMediaRepo,
		mockVideoService,
		mockCoreService,
		mockLogger,
	)

	mediaRequest := &model.MediaRequest{
		RequestID:    "test-request-id",
		UserID:       "test-user-id",
		Song:         "test-song",
		ProviderType: "test-provider",
	}

	mediaDetails := &model.MediaDetails{
		ID:           "test-video-id",
		Title:        "Test Title",
		DurationMs:   300000,
		URL:          "https://test-url.com",
		ThumbnailURL: "https://test-thumbnail.com",
		Provider:     "test-provider",
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	mockVideoService.On("GetMediaDetails", mock.Anything, mediaRequest.Song, mediaRequest.ProviderType).Return(mediaDetails, nil)

	mockMediaRepo.On("SaveMedia", mock.Anything, mock.MatchedBy(func(media *model.Media) bool {
		return media.VideoID == mediaDetails.ID && media.Metadata.Title == mediaDetails.Title
	})).Return(nil)

	mockCoreService.On("ProcessMedia", mock.Anything, mock.MatchedBy(func(media *model.Media) bool {
		return media.VideoID == mediaDetails.ID
	}), mediaRequest.UserID, mediaRequest.RequestID).Return(nil)

	err := processor.ProcessDownloadTask(ctx, mediaRequest)

	assert.NoError(t, err)
	mockMediaRepo.AssertExpectations(t)
	mockVideoService.AssertExpectations(t)
	mockCoreService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestMediaProcessor_ProcessRequest_GetMediaDetailsError(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockVideoService := new(MockVideoService)
	mockCoreService := new(MockCoreService)
	mockLogger := new(logger.MockLogger)

	processor := NewMediaProcessor(
		mockMediaRepo,
		mockVideoService,
		mockCoreService,
		mockLogger,
	)

	mediaRequest := &model.MediaRequest{
		RequestID:    "test-request-id",
		UserID:       "test-user-id",
		Song:         "test-song",
		ProviderType: "test-provider",
	}

	expectedError := errors.New("error obteniendo detalles")

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockVideoService.On("GetMediaDetails", mock.Anything, mediaRequest.Song, mediaRequest.ProviderType).Return(&model.MediaDetails{}, expectedError)

	err := processor.ProcessDownloadTask(ctx, mediaRequest)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockVideoService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestMediaProcessor_ProcessRequest_SaveMediaError(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockVideoService := new(MockVideoService)
	mockCoreService := new(MockCoreService)
	mockLogger := new(logger.MockLogger)

	processor := NewMediaProcessor(
		mockMediaRepo,
		mockVideoService,
		mockCoreService,
		mockLogger,
	)

	mediaRequest := &model.MediaRequest{
		RequestID:    "test-request-id",
		UserID:       "test-user-id",
		Song:         "test-song",
		ProviderType: "test-provider",
	}

	mediaDetails := &model.MediaDetails{
		ID:           "test-video-id",
		Title:        "Test Title",
		DurationMs:   300000,
		URL:          "https://test-url.com",
		ThumbnailURL: "https://test-thumbnail.com",
		Provider:     "test-provider",
	}

	expectedError := errors.New("error guardando media")

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockVideoService.On("GetMediaDetails", mock.Anything, mediaRequest.Song, mediaRequest.ProviderType).Return(mediaDetails, nil)

	mockMediaRepo.On("SaveMedia", mock.Anything, mock.MatchedBy(func(media *model.Media) bool {
		return media.VideoID == mediaDetails.ID && media.Metadata.Title == mediaDetails.Title
	})).Return(expectedError)

	err := processor.ProcessDownloadTask(ctx, mediaRequest)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockVideoService.AssertExpectations(t)
	mockMediaRepo.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestMediaProcessor_ProcessRequest_ProcessMediaError(t *testing.T) {
	ctx := context.Background()
	mockMediaRepo := new(MockMediaRepository)
	mockVideoService := new(MockVideoService)
	mockCoreService := new(MockCoreService)
	mockLogger := new(logger.MockLogger)

	processor := NewMediaProcessor(
		mockMediaRepo,
		mockVideoService,
		mockCoreService,
		mockLogger,
	)

	mediaRequest := &model.MediaRequest{
		RequestID:    "test-request-id",
		UserID:       "test-user-id",
		Song:         "test-song",
		ProviderType: "test-provider",
	}

	mediaDetails := &model.MediaDetails{
		ID:           "test-video-id",
		Title:        "Test Title",
		DurationMs:   300000,
		URL:          "https://test-url.com",
		ThumbnailURL: "https://test-thumbnail.com",
		Provider:     "test-provider",
	}

	expectedError := errors.New("error procesando media")

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockVideoService.On("GetMediaDetails", mock.Anything, mediaRequest.Song, mediaRequest.ProviderType).Return(mediaDetails, nil)

	mockMediaRepo.On("SaveMedia", mock.Anything, mock.MatchedBy(func(media *model.Media) bool {
		return media.VideoID == mediaDetails.ID && media.Metadata.Title == mediaDetails.Title
	})).Return(nil)

	mockCoreService.On("ProcessMedia", mock.Anything, mock.MatchedBy(func(media *model.Media) bool {
		return media.VideoID == mediaDetails.ID
	}), mediaRequest.UserID, mediaRequest.RequestID).Return(expectedError)

	err := processor.ProcessDownloadTask(ctx, mediaRequest)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockVideoService.AssertExpectations(t)
	mockMediaRepo.AssertExpectations(t)
	mockCoreService.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
