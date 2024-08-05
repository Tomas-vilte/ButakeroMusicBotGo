package service

import (
	"context"
	"encoding/json"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProcessSongMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var requestBody map[string]string
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		assert.NoError(t, err)
		assert.Equal(t, "Sean Paul - No Lie ft. Dua Lipa", requestBody["key"])
		assert.Equal(t, "Sean Paul - No Lie ft. Dua Lipa", requestBody["song"])

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]interface{}{
			"Type":         "youtube_provider",
			"Title":        "Sean Paul - No Lie ft. Dua Lipa",
			"URL":          "https://www.youtube.com/watch?v=GzU8KqOY8YA",
			"Playable":     true,
			"ThumbnailURL": "https://i.ytimg.com/vi/GzU8KqOY8YA/default.jpg",
			"Duration":     229000000000,
		})
		assert.NoError(t, err)
	}))
	defer server.Close()

	mockLogger := new(logging.MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	client := NewClient(server.URL, mockLogger)

	ctx := context.Background()
	result, err := client.ProcessSongMetadata(ctx, "Sean Paul - No Lie ft. Dua Lipa")

	// Verificar resultados
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "youtube_provider", result.Type)
	assert.Equal(t, "Sean Paul - No Lie ft. Dua Lipa", result.Title)
	assert.Equal(t, "https://www.youtube.com/watch?v=GzU8KqOY8YA", result.URL)
	assert.True(t, result.Playable)
	assert.Equal(t, "https://i.ytimg.com/vi/GzU8KqOY8YA/default.jpg", result.ThumbnailURL)
	assert.Equal(t, time.Duration(229000000000), result.Duration)

	mockLogger.AssertCalled(t, "Info", "Procesando metadata de la cancion a traves de API Gateway", mock.Anything)
}

func TestProcessSongMetadataError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	mockLogger := new(logging.MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	client := NewClient(server.URL, mockLogger)

	ctx := context.Background()
	result, err := client.ProcessSongMetadata(ctx, "Sean Paul - No Lie ft. Dua Lipa")

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "error en la solicitud a API Gateway: status code 500")

	mockLogger.AssertCalled(t, "Info", "Procesando metadata de la cancion a traves de API Gateway", mock.Anything)
	mockLogger.AssertCalled(t, "Error", "Error en la solicitud a API Gateway", mock.Anything)

}
