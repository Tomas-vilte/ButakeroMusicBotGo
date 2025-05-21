//go:build !integration

package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/errors_app"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func TestNewMediaAPIClient(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)

	cfg := AudioAPIClientConfig{
		BaseURL:         "http://localhost",
		Timeout:         time.Second * 5,
		MaxConnsPerHost: 5,
		MaxIdleConns:    10,
	}

	// act
	client, err := NewMediaAPIClient(cfg, mockLogger)

	// assert
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost", client.baseURL.String())
	assert.Equal(t, 5*time.Second, client.httpClient.Timeout)
}

func TestNewMediaAPIClient_InvalidURL(t *testing.T) {
	// Arrange
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)

	config := AudioAPIClientConfig{
		BaseURL: "://invalid-url",
	}

	// Act
	client, err := NewMediaAPIClient(config, mockLogger)

	// Assert
	require.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), "La URL base no es válida")
}

func TestGetMediaByID_Success(t *testing.T) {
	// Arrange
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/media", r.URL.Path)
		assert.Equal(t, "test123", r.URL.Query().Get("video_id"))

		media := &model.Media{
			Status: "completed",
			Metadata: model.Metadata{
				Title:        "Test Song",
				DurationMs:   180000,
				URL:          "https://example.com/test123",
				ThumbnailURL: "https://example.com/thumbnail.jpg",
				Platform:     "youtube",
			},
			FileData: model.FileData{
				FilePath: "/path/to/file",
				FileSize: "10MB",
				FileType: "mp3",
			},
			Success:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		response := model.MediaResponse{
			Data:    media,
			Success: true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			return
		}
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := &MediaAPIClient{
		baseURL:    baseURL,
		logger:     mockLogger,
		httpClient: server.Client(),
	}

	// Act
	result, err := client.GetMediaByID(context.Background(), "test123")

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Test Song", result.Metadata.Title)
	assert.Equal(t, "completed", result.Status)
	assert.Equal(t, true, result.Success)
}

func TestGetMediaByID_NotFound(t *testing.T) {
	// Arrange
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/media", r.URL.Path)
		assert.Equal(t, "nonexistent", r.URL.Query().Get("video_id"))

		errorResp := model.ErrorResponse{
			Error: model.ErrorDetail{
				Code:    "media_not_found",
				Message: "Media no encontrado",
				VideoID: "nonexistent",
			},
			Success: false,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		err := json.NewEncoder(w).Encode(errorResp)
		if err != nil {
			return
		}
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := &MediaAPIClient{
		baseURL:    baseURL,
		logger:     mockLogger,
		httpClient: server.Client(),
	}

	// Act
	result, err := client.GetMediaByID(context.Background(), "nonexistent")

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "No se pudo obtener la canción")
}

func TestGetMediaByID_ServerError(t *testing.T) {
	// Arrange
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := &MediaAPIClient{
		baseURL:    baseURL,
		logger:     mockLogger,
		httpClient: server.Client(),
	}

	// Act
	result, err := client.GetMediaByID(context.Background(), "test123")

	// Assert
	require.Error(t, err)
	assert.Nil(t, result)
	var appErr *errors_app.AppError
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, errors_app.ErrCodeInternalError, appErr.Code)
}

func TestSearchMediaByTitle_Success(t *testing.T) {
	// Arrange
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/media/search", r.URL.Path)
		assert.Equal(t, "test song", r.URL.Query().Get("title"))

		media1 := &model.Media{
			Status: "completed",
			Metadata: model.Metadata{
				Title:        "Test Song 1",
				DurationMs:   180000,
				URL:          "https://example.com/test1",
				ThumbnailURL: "https://example.com/thumbnail1.jpg",
				Platform:     "youtube",
			},
			Success: true,
		}

		media2 := &model.Media{
			Status: "completed",
			Metadata: model.Metadata{
				Title:        "Test Song 2",
				DurationMs:   240000,
				URL:          "https://example.com/test2",
				ThumbnailURL: "https://example.com/thumbnail2.jpg",
				Platform:     "youtube",
			},
			Success: true,
		}

		response := model.MediaListResponse{
			Data:    []*model.Media{media1, media2},
			Success: true,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			return
		}
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := &MediaAPIClient{
		baseURL:    baseURL,
		logger:     mockLogger,
		httpClient: server.Client(),
	}

	// Act
	results, err := client.SearchMediaByTitle(context.Background(), "test song")

	// Assert
	require.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Equal(t, "Test Song 1", results[0].Metadata.Title)
	assert.Equal(t, "Test Song 2", results[1].Metadata.Title)
}

func TestSearchMediaByTitle_NoResults(t *testing.T) {
	// Arrange
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/media/search", r.URL.Path)
		assert.Equal(t, "nonexistent", r.URL.Query().Get("title"))

		response := model.MediaListResponse{
			Data:    []*model.Media{},
			Success: false,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			return
		}
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := &MediaAPIClient{
		baseURL:    baseURL,
		logger:     mockLogger,
		httpClient: server.Client(),
	}

	// Act
	results, err := client.SearchMediaByTitle(context.Background(), "nonexistent")

	// Assert
	require.Error(t, err)
	assert.Nil(t, results)
	var appErr *errors_app.AppError
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, errors_app.ErrCodeInternalError, appErr.Code)
}

func TestSearchMediaByTitle_ServerError(t *testing.T) {
	// Arrange
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)
	client := &MediaAPIClient{
		baseURL:    baseURL,
		logger:     mockLogger,
		httpClient: server.Client(),
	}

	// Act
	results, err := client.SearchMediaByTitle(context.Background(), "test song")

	// Assert
	require.Error(t, err)
	assert.Nil(t, results)
	var appErr *errors_app.AppError
	ok := errors.As(err, &appErr)
	assert.True(t, ok)
	assert.Equal(t, errors_app.ErrCodeInternalError, appErr.Code)
}
