//go:build !integration

package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model/queue"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestGetSongFromAPI_WithValidURL(t *testing.T) {
	// arrange
	ctx := context.Background()
	mockMediaClient := new(MockMediaClient)
	mockPublisher := new(MockSongDownloadRequestPublisher)
	mockSubscriber := new(MockSongDownloadEventSubscriber)
	mockLogger := new(logging.MockLogger)

	expectedMedia := &model.Media{
		Metadata: model.Metadata{
			Title:        "Sample Title",
			DurationMs:   300000,
			Platform:     "youtube",
			ThumbnailURL: "https://example.com/thumbnail.jpg",
			URL:          "https://youtube.com/watch?v=sample",
		},
		FileData: model.FileData{
			FilePath: "/path/to/file.mp3",
		},
	}

	url := "https://www.youtube.com/watch?v=sample"
	mockMediaClient.On("GetMediaByID", ctx, "sample").Return(expectedMedia, nil)

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockSubscriber.On("DownloadEventsChannel").Return(make(chan *queue.DownloadStatusMessage))

	service := NewSongService(mockMediaClient, mockPublisher, mockSubscriber, mockLogger)

	// act
	result, err := service.GetSongFromAPI(ctx, url)

	// assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedMedia.Metadata.Title, result.TitleTrack)
	assert.Equal(t, expectedMedia.Metadata.DurationMs, result.DurationMs)
	assert.Equal(t, expectedMedia.Metadata.Platform, result.Platform)
	assert.Equal(t, expectedMedia.Metadata.ThumbnailURL, result.ThumbnailURL)
	assert.Equal(t, expectedMedia.Metadata.URL, result.URL)
	assert.Equal(t, expectedMedia.FileData.FilePath, result.FilePath)

	mockMediaClient.AssertExpectations(t)
}

func TestGetSongFromAPI_WithValidTitle(t *testing.T) {
	// arrange
	ctx := context.Background()
	mockMediaClient := new(MockMediaClient)
	mockPublisher := new(MockSongDownloadRequestPublisher)
	mockSubscriber := new(MockSongDownloadEventSubscriber)
	mockLogger := new(logging.MockLogger)

	expectedMedia := &model.Media{
		Metadata: model.Metadata{
			Title:        "Sample Title",
			DurationMs:   300000,
			Platform:     "youtube",
			ThumbnailURL: "https://example.com/thumbnail.jpg",
			URL:          "https://youtube.com/watch?v=sample",
		},
		FileData: model.FileData{
			FilePath: "/path/to/file.mp3",
		},
	}

	title := "Sample Song Title"
	mockMediaClient.On("SearchMediaByTitle", ctx, title).Return([]*model.Media{expectedMedia}, nil)
	downloadEventsChan := make(chan *queue.DownloadStatusMessage, 1)
	mockSubscriber.On("DownloadEventsChannel").Return(downloadEventsChan)

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	service := NewSongService(mockMediaClient, mockPublisher, mockSubscriber, mockLogger)

	// act
	result, err := service.GetSongFromAPI(ctx, title)

	// assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedMedia.Metadata.Title, result.TitleTrack)
	assert.Equal(t, expectedMedia.Metadata.DurationMs, result.DurationMs)
	assert.Equal(t, expectedMedia.Metadata.Platform, result.Platform)
	assert.Equal(t, expectedMedia.Metadata.ThumbnailURL, result.ThumbnailURL)
	assert.Equal(t, expectedMedia.Metadata.URL, result.URL)
	assert.Equal(t, expectedMedia.FileData.FilePath, result.FilePath)

	mockMediaClient.AssertExpectations(t)
}

func TestGetSongFromAPI_WithInvalidURL(t *testing.T) {
	// arrange
	ctx := context.Background()
	mockMediaClient := new(MockMediaClient)
	mockPublisher := new(MockSongDownloadRequestPublisher)
	mockSubscriber := new(MockSongDownloadEventSubscriber)
	mockLogger := new(logging.MockLogger)

	url := "https://www.youtube.com/watch?v=nonexistent"
	mockMediaClient.On("GetMediaByID", ctx, "nonexistent").Return((*model.Media)(nil), errors.New("not found"))

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	service := NewSongService(mockMediaClient, mockPublisher, mockSubscriber, mockLogger)
	downloadEventsChan := make(chan *queue.DownloadStatusMessage, 1)
	mockSubscriber.On("DownloadEventsChannel").Return(downloadEventsChan)

	// act
	result, err := service.GetSongFromAPI(ctx, url)

	// assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "canción no encontrada en la API")

	mockMediaClient.AssertExpectations(t)
}

func TestGetSongFromAPI_WithInvalidTitle(t *testing.T) {
	// arrange
	ctx := context.Background()
	mockMediaClient := new(MockMediaClient)
	mockPublisher := new(MockSongDownloadRequestPublisher)
	mockSubscriber := new(MockSongDownloadEventSubscriber)
	mockLogger := new(logging.MockLogger)

	title := "Nonexistent Song Title"
	mockMediaClient.On("SearchMediaByTitle", ctx, title).Return([]*model.Media{}, nil)

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	downloadEventsChan := make(chan *queue.DownloadStatusMessage, 1)
	mockSubscriber.On("DownloadEventsChannel").Return(downloadEventsChan)

	service := NewSongService(mockMediaClient, mockPublisher, mockSubscriber, mockLogger)

	// act
	result, err := service.GetSongFromAPI(ctx, title)

	// assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "canción no encontrada en la API")

	mockMediaClient.AssertExpectations(t)
}

func TestDownloadSongViaQueue_Success(t *testing.T) {
	// arrange
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	mockMediaClient := new(MockMediaClient)
	mockPublisher := new(MockSongDownloadRequestPublisher)
	mockSubscriber := new(MockSongDownloadEventSubscriber)
	mockLogger := new(logging.MockLogger)

	userID := "user123"
	input := "https://www.youtube.com/watch?v=sample"
	providerType := "youtube"

	var capturedRequestID string
	mockPublisher.On("PublishDownloadRequest", mock.Anything, mock.MatchedBy(func(req *queue.DownloadRequestMessage) bool {
		capturedRequestID = req.RequestID
		return req.UserID == userID && req.Song == input && req.ProviderType == providerType
	})).Return(nil)

	downloadEventsChan := make(chan *queue.DownloadStatusMessage, 1)
	mockSubscriber.On("DownloadEventsChannel").Return(downloadEventsChan)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	service := NewSongService(mockMediaClient, mockPublisher, mockSubscriber, mockLogger)

	go func() {
		time.Sleep(100 * time.Millisecond)
		expectedResponse := &queue.DownloadStatusMessage{
			RequestID: capturedRequestID,
			Status:    "success",
			VideoID:   "sample",
			PlatformMetadata: queue.SongMetadata{
				Title:        "Sample Title",
				DurationMs:   300000,
				ThumbnailURL: "https://example.com/thumbnail.jpg",
				Platform:     "youtube",
				URL:          "https://youtube.com/watch?v=sample",
			},
			FileData: queue.FileData{
				FilePath: "/path/to/file.mp3",
			},
		}
		downloadEventsChan <- expectedResponse
	}()

	// act
	result, err := service.DownloadSongViaQueue(ctx, userID, input, providerType)

	// assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Sample Title", result.TitleTrack)
	assert.Equal(t, int64(300000), result.DurationMs)
	assert.Equal(t, "youtube", result.Platform)
	assert.Equal(t, "https://example.com/thumbnail.jpg", result.ThumbnailURL)
	assert.Equal(t, "https://youtube.com/watch?v=sample", result.URL)
	assert.Equal(t, "/path/to/file.mp3", result.FilePath)

	mockPublisher.AssertExpectations(t)
	mockSubscriber.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestDownloadSongViaQueue_PublishError(t *testing.T) {
	// arrange
	ctx := context.Background()
	mockMediaClient := new(MockMediaClient)
	mockPublisher := new(MockSongDownloadRequestPublisher)
	mockSubscriber := new(MockSongDownloadEventSubscriber)
	mockLogger := new(logging.MockLogger)

	userID := "user123"
	input := "https://www.youtube.com/watch?v=sample"
	providerType := "youtube"
	publishError := errors.New("queue publish error")

	mockPublisher.On("PublishDownloadRequest", mock.Anything, mock.MatchedBy(func(req *queue.DownloadRequestMessage) bool {
		return req.UserID == userID && req.Song == input && req.ProviderType == providerType
	})).Return(publishError)

	mockSubscriber.On("DownloadEventsChannel").Return(make(chan *queue.DownloadStatusMessage))

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	service := NewSongService(mockMediaClient, mockPublisher, mockSubscriber, mockLogger)

	// act
	result, err := service.DownloadSongViaQueue(ctx, userID, input, providerType)

	// assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error al solicitar la descarga")

	mockPublisher.AssertExpectations(t)
}

func TestDownloadSongViaQueue_DownloadFailure(t *testing.T) {
	// arrange
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	mockMediaClient := new(MockMediaClient)
	mockPublisher := new(MockSongDownloadRequestPublisher)
	mockSubscriber := new(MockSongDownloadEventSubscriber)
	mockLogger := new(logging.MockLogger)

	userID := "user123"
	input := "https://www.youtube.com/watch?v=sample"
	providerType := "youtube"

	var capturedRequestID string
	mockPublisher.On("PublishDownloadRequest", mock.Anything, mock.MatchedBy(func(req *queue.DownloadRequestMessage) bool {
		capturedRequestID = req.RequestID
		return req.UserID == userID && req.Song == input && req.ProviderType == providerType
	})).Return(nil)

	downloadEventsChan := make(chan *queue.DownloadStatusMessage, 1)
	mockSubscriber.On("DownloadEventsChannel").Return(downloadEventsChan)

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	service := NewSongService(mockMediaClient, mockPublisher, mockSubscriber, mockLogger)

	go func() {
		time.Sleep(100 * time.Millisecond)
		errorResponse := &queue.DownloadStatusMessage{
			RequestID: capturedRequestID,
			Status:    "error",
			Message:   "Failed to download the song",
		}
		downloadEventsChan <- errorResponse
	}()

	// act
	result, err := service.DownloadSongViaQueue(ctx, userID, input, providerType)

	// assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error en la descarga")

	mockPublisher.AssertExpectations(t)
	mockSubscriber.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestDownloadSongViaQueue_Timeout(t *testing.T) {
	// arrange
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	mockMediaClient := new(MockMediaClient)
	mockPublisher := new(MockSongDownloadRequestPublisher)
	mockSubscriber := new(MockSongDownloadEventSubscriber)
	mockLogger := new(logging.MockLogger)

	userID := "user123"
	input := "https://www.youtube.com/watch?v=sample"
	providerType := "youtube"

	mockPublisher.On("PublishDownloadRequest", mock.Anything, mock.MatchedBy(func(req *queue.DownloadRequestMessage) bool {
		return req.UserID == userID && req.Song == input && req.ProviderType == providerType
	})).Return(nil)

	downloadEventsChan := make(chan *queue.DownloadStatusMessage, 1)
	mockSubscriber.On("DownloadEventsChannel").Return(downloadEventsChan)

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	service := NewSongService(mockMediaClient, mockPublisher, mockSubscriber, mockLogger)

	// act
	result, err := service.DownloadSongViaQueue(ctx, userID, input, providerType)

	// assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "tiempo de espera agotado")

	mockPublisher.AssertExpectations(t)
	mockSubscriber.AssertExpectations(t)
}

func TestGetOrDownloadSong_APISuccess(t *testing.T) {
	// arrange
	ctx := context.Background()
	mockMediaClient := new(MockMediaClient)
	mockPublisher := new(MockSongDownloadRequestPublisher)
	mockSubscriber := new(MockSongDownloadEventSubscriber)
	mockLogger := new(logging.MockLogger)

	userID := "user123"
	input := "https://www.youtube.com/watch?v=sample"
	providerType := "youtube"

	expectedMedia := &model.Media{
		Metadata: model.Metadata{
			Title:        "Sample Title",
			DurationMs:   300000,
			Platform:     "youtube",
			ThumbnailURL: "https://example.com/thumbnail.jpg",
			URL:          "https://youtube.com/watch?v=sample",
		},
		FileData: model.FileData{
			FilePath: "/path/to/file.mp3",
		},
	}

	mockSubscriber.On("DownloadEventsChannel").Return(make(chan *queue.DownloadStatusMessage))
	mockMediaClient.On("GetMediaByID", ctx, "sample").Return(expectedMedia, nil)

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	service := NewSongService(mockMediaClient, mockPublisher, mockSubscriber, mockLogger)

	// act
	result, err := service.GetOrDownloadSong(ctx, userID, input, providerType)

	// assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedMedia.Metadata.Title, result.TitleTrack)
	assert.Equal(t, expectedMedia.Metadata.DurationMs, result.DurationMs)
	assert.Equal(t, expectedMedia.Metadata.Platform, result.Platform)
	assert.Equal(t, expectedMedia.Metadata.ThumbnailURL, result.ThumbnailURL)
	assert.Equal(t, expectedMedia.Metadata.URL, result.URL)
	assert.Equal(t, expectedMedia.FileData.FilePath, result.FilePath)

	mockMediaClient.AssertExpectations(t)
	mockPublisher.AssertNotCalled(t, "PublishDownloadRequest")
}

func TestGetOrDownloadSong_APIFailure_DownloadSuccess(t *testing.T) {
	// arrange
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	mockMediaClient := new(MockMediaClient)
	mockPublisher := new(MockSongDownloadRequestPublisher)
	mockSubscriber := new(MockSongDownloadEventSubscriber)
	mockLogger := new(logging.MockLogger)

	userID := "user123"
	input := "https://www.youtube.com/watch?v=sample"
	providerType := "youtube"

	mockMediaClient.On("GetMediaByID", ctx, "sample").Return((*model.Media)(nil), fmt.Errorf("API error"))

	var capturedRequestID string
	mockPublisher.On("PublishDownloadRequest", mock.Anything, mock.MatchedBy(func(req *queue.DownloadRequestMessage) bool {
		capturedRequestID = req.RequestID
		return req.UserID == userID && req.Song == input && req.ProviderType == providerType
	})).Return(nil)

	downloadEventsChan := make(chan *queue.DownloadStatusMessage, 1)
	mockSubscriber.On("DownloadEventsChannel").Return(downloadEventsChan)

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	service := NewSongService(mockMediaClient, mockPublisher, mockSubscriber, mockLogger)

	go func() {
		time.Sleep(100 * time.Millisecond)
		expectedResponse := &queue.DownloadStatusMessage{
			RequestID: capturedRequestID,
			Status:    "success",
			VideoID:   "sample",
			PlatformMetadata: queue.SongMetadata{
				Title:        "Sample Title",
				DurationMs:   300000,
				ThumbnailURL: "https://example.com/thumbnail.jpg",
				Platform:     "youtube",
				URL:          "https://youtube.com/watch?v=sample",
			},
			FileData: queue.FileData{
				FilePath: "/path/to/file.mp3",
			},
		}
		downloadEventsChan <- expectedResponse
	}()

	// act
	result, err := service.GetOrDownloadSong(ctx, userID, input, providerType)

	// assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Sample Title", result.TitleTrack)
	assert.Equal(t, int64(300000), result.DurationMs)
	assert.Equal(t, "youtube", result.Platform)
	assert.Equal(t, "https://example.com/thumbnail.jpg", result.ThumbnailURL)
	assert.Equal(t, "https://youtube.com/watch?v=sample", result.URL)
	assert.Equal(t, "/path/to/file.mp3", result.FilePath)

	mockMediaClient.AssertExpectations(t)
	mockPublisher.AssertExpectations(t)
	mockSubscriber.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
