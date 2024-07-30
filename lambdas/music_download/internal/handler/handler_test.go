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
		mockSQSClient := new(MockSQSClient)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockSQSClient)
		songEvent := SongEvent{
			Song: "Test Song",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

		mockYouTubeClient.On("SearchYouTubeVideoID", mock.Anything, "Test Song").Return("testVideoID", nil)
		mockYouTubeClient.On("LookupSongs", mock.Anything, "testVideoID").Return([]*types.Song{{Title: "Test Song"}}, nil)
		mockDownloader.On("DownloadSong", "https://www.youtube.com/watch?v=testVideoID", "audio_input_raw/Test Song.m4a").Return(nil)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return(nil)
		mockSQSClient.On("SendMessage", mock.Anything, mock.Anything).Return(nil)

		response, err := h.HandleEvent(context.Background(), event)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, response.StatusCode)

		var responseSong types.Song
		err = json.Unmarshal([]byte(response.Body), &responseSong)
		assert.NoError(t, err)
		assert.Equal(t, "Test Song", responseSong.Title)

		mockLogger.AssertExpectations(t)
		mockYouTubeClient.AssertExpectations(t)
		mockDownloader.AssertExpectations(t)
	})

	t.Run("Invalid event body", func(t *testing.T) {
		mockDownloader := new(MockDownloader)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)
		mockYouTubeClient := new(MockSongLooker)
		mockSQSClient := new(MockSQSClient)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockSQSClient)
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
		mockSQSClient := new(MockSQSClient)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockSQSClient)
		songEvent := SongEvent{
			Song: "Test Song",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

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
		mockSQSClient := new(MockSQSClient)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockSQSClient)
		songEvent := SongEvent{
			Song: "Test Song",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

		mockLogger.On("Info", mock.Anything, mock.Anything).Return()
		mockYouTubeClient.On("SearchYouTubeVideoID", mock.Anything, "Test Song").Return("testVideoID", nil)
		mockYouTubeClient.On("LookupSongs", mock.Anything, "testVideoID").Return([]*types.Song{{Title: "Test Song"}}, nil)
		mockDownloader.On("DownloadSong", "https://www.youtube.com/watch?v=testVideoID", "audio_input_raw/Test Song.m4a").Return(errors.New("download error"))
		mockLogger.On("Error", "Error al descargar la canción", mock.Anything).Return()

		response, err := h.HandleEvent(context.Background(), event)

		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)
		assert.Contains(t, response.Body, "Error al descargar la canción")

		mockLogger.AssertExpectations(t)
		mockDownloader.AssertExpectations(t)
	})

	t.Run("Error obtaining video details", func(t *testing.T) {
		mockDownloader := new(MockDownloader)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)
		mockYouTubeClient := new(MockSongLooker)
		mockSQSClient := new(MockSQSClient)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockSQSClient)
		songEvent := SongEvent{
			Song: "Test Song",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

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
		mockSQSClient := new(MockSQSClient)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockSQSClient)
		songEvent := SongEvent{
			Song: "Test Song",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

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

	t.Run("SQS send message error", func(t *testing.T) {
		mockDownloader := new(MockDownloader)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)
		mockYouTubeClient := new(MockSongLooker)
		mockSQSClient := new(MockSQSClient)

		h := NewHandler(mockDownloader, mockUploader, mockLogger, mockYouTubeClient, mockSQSClient)
		songEvent := SongEvent{
			Song: "Test Song",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

		mockYouTubeClient.On("SearchYouTubeVideoID", mock.Anything, "Test Song").Return("testVideoID", nil)
		mockYouTubeClient.On("LookupSongs", mock.Anything, "testVideoID").Return([]*types.Song{{Title: "Test Song"}}, nil)
		mockDownloader.On("DownloadSong", "https://www.youtube.com/watch?v=testVideoID", "audio_input_raw/Test Song.m4a").Return(nil)
		mockSQSClient.On("SendMessage", mock.Anything, mock.Anything).Return(errors.New("sqs send error"))
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()
		mockLogger.On("Error", "Error al enviar el mensaje a SQS", mock.Anything).Return()

		response, err := h.HandleEvent(context.Background(), event)

		assert.Error(t, err)
		assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

		mockLogger.AssertExpectations(t)
		mockYouTubeClient.AssertExpectations(t)
		mockDownloader.AssertExpectations(t)
		mockSQSClient.AssertExpectations(t)
	})
}
