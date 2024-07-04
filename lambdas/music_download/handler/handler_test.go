package handler

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestHandleEvent(t *testing.T) {

	t.Run("Successful event handling", func(t *testing.T) {
		mockDownloader := new(MockDownloader)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)

		h := NewHandler(mockDownloader, mockUploader, mockLogger)
		songEvent := SongEvent{
			URL: "https://example.com/song",
			Key: "song.m4a",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

		mockLogger.On("Info", mock.Anything, mock.Anything)
		mockDownloader.On("DownloadSong", songEvent.URL, "audio_input_raw/song.m4a").Return(nil)
		mockLogger.On("Info", "Canción procesada exitosamente", mock.Anything)

		response, err := h.HandleEvent(context.Background(), event)

		assert.NoError(t, err)
		assert.Equal(t, 200, response.StatusCode)
		assert.Equal(t, "Canción procesada exitosamente", response.Body)

		mockLogger.AssertExpectations(t)
		mockDownloader.AssertExpectations(t)
	})

	t.Run("Invalid event body", func(t *testing.T) {
		mockDownloader := new(MockDownloader)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)

		h := NewHandler(mockDownloader, mockUploader, mockLogger)
		event := events.APIGatewayProxyRequest{
			Body: "invalid json",
		}

		mockLogger.On("Info", mock.Anything, mock.Anything)
		mockLogger.On("Error", "Error al parsear el evento", mock.Anything)

		response, err := h.HandleEvent(context.Background(), event)

		assert.Error(t, err)
		assert.Equal(t, 400, response.StatusCode)
		assert.Contains(t, response.Body, "Error al parsear el evento")

		mockLogger.AssertExpectations(t)
	})
	t.Run("Download error", func(t *testing.T) {
		mockDownloader := new(MockDownloader)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)

		h := NewHandler(mockDownloader, mockUploader, mockLogger)
		songEvent := SongEvent{
			URL: "https://example.com/song",
			Key: "song.m4a",
		}
		eventBody, _ := json.Marshal(songEvent)

		event := events.APIGatewayProxyRequest{
			Body: string(eventBody),
		}

		mockLogger.On("Info", mock.Anything, mock.Anything)
		mockDownloader.On("DownloadSong", songEvent.URL, "audio_input_raw/song.m4a").Return(errors.New("download error"))
		mockLogger.On("Error", "Error al descargar la canción", mock.Anything)

		response, err := h.HandleEvent(context.Background(), event)

		assert.Error(t, err)
		assert.Equal(t, 500, response.StatusCode)
		assert.Contains(t, response.Body, "Error al descargar la canción")

		mockLogger.AssertExpectations(t)
		mockDownloader.AssertExpectations(t)
	})
}
