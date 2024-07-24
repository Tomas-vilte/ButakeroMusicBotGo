package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/types"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"testing"
)

func TestHandleEvent(t *testing.T) {

	t.Run("Successful event handling", func(t *testing.T) {
		mockDownloader := new(MockDownloader)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)
		mockYouTubeClient := new(MockSongLooker)
		mockCache := new(CacheMock)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockCache)
		songEvent := SongEvent{
			Song: "Test Song",
			Key:  "testKey",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

		mockCache.On("GetSong", mock.Anything, "testKey").Return(nil, nil)
		mockYouTubeClient.On("SearchYouTubeVideoID", mock.Anything, "Test Song").Return("testVideoID", nil)
		mockYouTubeClient.On("LookupSongs", mock.Anything, "testVideoID").Return([]*types.Song{{Title: "Test Song"}}, nil)
		mockDownloader.On("DownloadSong", "https://www.youtube.com/watch?v=testVideoID", "audio_input_raw/testKey.m4a").Return(nil)
		mockCache.On("SetSong", mock.Anything, "testKey", mock.AnythingOfType("*types.Song")).Return(nil)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return(nil)

		response, err := h.HandleEvent(context.Background(), event)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		var responseSong types.Song
		err = json.Unmarshal([]byte(response.Body), &responseSong)
		assert.NoError(t, err)
		assert.Equal(t, "Test Song", responseSong.Title)

		mockCache.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
		mockYouTubeClient.AssertExpectations(t)
		mockDownloader.AssertExpectations(t)
	})

	t.Run("Invalid event body", func(t *testing.T) {
		mockDownloader := new(MockDownloader)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)
		mockYouTubeClient := new(MockSongLooker)
		mockCache := new(CacheMock)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockCache)
		event := events.APIGatewayProxyRequest{
			Body: "invalid json",
		}

		mockLogger.On("Info", mock.Anything, mock.Anything).Return()
		mockLogger.On("Error", "Error al parsear el evento", mock.Anything).Return()

		response, err := h.HandleEvent(context.Background(), event)

		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.Contains(t, response.Body, "Error al parsear el evento")

		mockLogger.AssertExpectations(t)
	})

	t.Run("YouTube search error", func(t *testing.T) {
		mockDownloader := new(MockDownloader)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)
		mockYouTubeClient := new(MockSongLooker)
		mockCache := new(CacheMock)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockCache)
		songEvent := SongEvent{
			Song: "Test Song",
			Key:  "testKey",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

		mockCache.On("GetSong", mock.Anything, "testKey").Return(nil, nil)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()
		mockYouTubeClient.On("SearchYouTubeVideoID", mock.Anything, "Test Song").Return("", errors.New("search error"))
		mockLogger.On("Error", "Error al buscar el ID del video en YouTube", mock.Anything, mock.Anything).Return()

		response, err := h.HandleEvent(context.Background(), event)

		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.Contains(t, response.Body, "Error al buscar el ID del video")

		mockLogger.AssertExpectations(t)
		mockYouTubeClient.AssertExpectations(t)
	})

	t.Run("Download error", func(t *testing.T) {
		mockDownloader := new(MockDownloader)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)
		mockYouTubeClient := new(MockSongLooker)
		mockCache := new(CacheMock)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockCache)
		songEvent := SongEvent{
			Song: "Test Song",
			Key:  "testKey",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

		mockCache.On("GetSong", mock.Anything, "testKey").Return(nil, nil)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()
		mockYouTubeClient.On("SearchYouTubeVideoID", mock.Anything, "Test Song").Return("testVideoID", nil)
		mockYouTubeClient.On("LookupSongs", mock.Anything, "testVideoID").Return([]*types.Song{{Title: "Test Song"}}, nil)
		mockDownloader.On("DownloadSong", "https://www.youtube.com/watch?v=testVideoID", "audio_input_raw/testKey.m4a").Return(errors.New("download error"))
		mockLogger.On("Error", "Error al descargar la canción", mock.Anything).Return()

		response, err := h.HandleEvent(context.Background(), event)

		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Contains(t, response.Body, "Error al descargar la canción")

		mockLogger.AssertExpectations(t)
		mockYouTubeClient.AssertExpectations(t)
		mockDownloader.AssertExpectations(t)
	})

	t.Run("Error obtaining video details", func(t *testing.T) {
		mockDownloader := new(MockDownloader)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)
		mockYouTubeClient := new(MockSongLooker)
		mockCache := new(CacheMock)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockCache)
		songEvent := SongEvent{
			Song: "Test Song",
			Key:  "testKey",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

		mockCache.On("GetSong", mock.Anything, "testKey").Return(nil, nil)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()
		mockYouTubeClient.On("SearchYouTubeVideoID", mock.Anything, "Test Song").Return("testVideoID", nil)
		mockYouTubeClient.On("LookupSongs", mock.Anything, "testVideoID").Return([]*types.Song{}, errors.New("error obtaining video details"))
		mockLogger.On("Error", "Error al obtener detalles del video", mock.Anything).Return()

		response, err := h.HandleEvent(context.Background(), event)

		assert.Error(t, err)
		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
		assert.Contains(t, response.Body, "Error al obtener detalles del video")

		mockLogger.AssertExpectations(t)
		mockYouTubeClient.AssertExpectations(t)
	})

	t.Run("No video details found", func(t *testing.T) {
		mockDownloader := new(MockDownloader)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)
		mockYouTubeClient := new(MockSongLooker)
		mockCache := new(CacheMock)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockCache)
		songEvent := SongEvent{
			Song: "Test Song",
			Key:  "testKey",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

		mockCache.On("GetSong", mock.Anything, "testKey").Return(nil, nil)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()
		mockYouTubeClient.On("SearchYouTubeVideoID", mock.Anything, "Test Song").Return("testVideoID", nil)
		mockYouTubeClient.On("LookupSongs", mock.Anything, "testVideoID").Return([]*types.Song{}, nil)
		mockLogger.On("Error", "No se encontraron detalles del video", mock.Anything).Return()

		response, err := h.HandleEvent(context.Background(), event)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, response.StatusCode)
		assert.Equal(t, "No se encontraron detalles del video", response.Body)

		mockLogger.AssertExpectations(t)
		mockYouTubeClient.AssertExpectations(t)
	})

	t.Run("Cache retrieval error", func(t *testing.T) {
		mockDownloader := new(MockDownloader)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)
		mockYouTubeClient := new(MockSongLooker)
		mockCache := new(CacheMock)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockCache)
		songEvent := SongEvent{
			Song: "Test Song",
			Key:  "testKey",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

		mockCache.On("GetSong", mock.Anything, "testKey").Return(nil, errors.New("cache error"))
		mockLogger.On("Error", "Error al obtener del cache", mock.Anything).Return()
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()

		response, err := h.HandleEvent(context.Background(), event)

		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Contains(t, response.Body, "Error al obtener del cache")

		mockLogger.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("Song found in cache", func(t *testing.T) {
		mockDownloader := new(MockDownloader)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)
		mockYouTubeClient := new(MockSongLooker)
		mockCache := new(CacheMock)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockCache)
		songEvent := SongEvent{
			Song: "Test Song",
			Key:  "testKey",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

		cachedSong := &types.Song{Title: "Test Song"}
		mockCache.On("GetSong", mock.Anything, "testKey").Return(cachedSong, nil)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()

		response, err := h.HandleEvent(context.Background(), event)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		var responseSong types.Song
		err = json.Unmarshal([]byte(response.Body), &responseSong)
		assert.NoError(t, err)
		assert.Equal(t, "Test Song", responseSong.Title)

		mockLogger.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("Cache save error", func(t *testing.T) {
		mockDownloader := new(MockDownloader)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)
		mockYouTubeClient := new(MockSongLooker)
		mockCache := new(CacheMock)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockCache)
		songEvent := SongEvent{
			Song: "Test Song",
			Key:  "testKey",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

		mockLogger.On("Info", mock.Anything, mock.Anything).Return()
		mockYouTubeClient.On("SearchYouTubeVideoID", mock.Anything, "Test Song").Return("testVideoID", nil)
		mockYouTubeClient.On("LookupSongs", mock.Anything, "testVideoID").Return([]*types.Song{{Title: "Test Song"}}, nil)
		mockDownloader.On("DownloadSong", "https://www.youtube.com/watch?v=testVideoID", "audio_input_raw/testKey.m4a").Return(nil)
		mockCache.On("SetSong", mock.Anything, "testKey", mock.AnythingOfType("*types.Song")).Return(errors.New("cache save error"))

		mockCache.On("GetSong", mock.Anything, "testKey").Return(nil, nil)

		mockLogger.On("Error", "Error al guardar en cache", mock.Anything).Return()

		response, err := h.HandleEvent(context.Background(), event)

		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Contains(t, response.Body, "Error al guardar en cache")

		mockLogger.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

}
