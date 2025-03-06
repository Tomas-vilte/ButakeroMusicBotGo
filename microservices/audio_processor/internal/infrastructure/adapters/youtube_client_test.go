//go:build integration

package adapters

import (
	"context"
	"encoding/json"
	"errors"
	errorsApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestYouTubeClient_GetVideoDetails(t *testing.T) {
	t.Run("debe retornar detalles del video cuando la respuesta es válida", func(t *testing.T) {
		// Arrange
		expectedID := "dQw4w9WgXcQ"
		publishedDate := time.Now().Format(time.RFC3339)
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Contains(t, r.URL.String(), "videos?part=snippet,contentDetails")
			assert.Contains(t, r.URL.String(), "id="+expectedID)

			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"id": expectedID,
						"snippet": map[string]interface{}{
							"title":        "Test Video",
							"description":  "Test Description",
							"channelTitle": "Test Channel",
							"publishedAt":  publishedDate,
							"thumbnails": map[string]interface{}{
								"default": map[string]interface{}{
									"url": "http://test.com/thumb.jpg",
								},
							},
						},
						"contentDetails": map[string]interface{}{
							"duration": "PT5M30S",
						},
					},
				},
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Fatal(err)
			}
		}))
		defer ts.Close()

		mockLogger := new(logger.MockLogger)
		mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
		client := NewYouTubeClient("test-key", mockLogger)
		client.BaseURL = ts.URL

		// Act
		details, err := client.GetVideoDetails(context.Background(), expectedID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedID, details.ID)
		assert.Equal(t, "Test Video", details.Title)
		assert.Equal(t, "Test Description", details.Description)
		assert.Equal(t, "Test Channel", details.Creator)
		assert.Equal(t, "PT5M30S", details.Duration)
		assert.Equal(t, "http://test.com/thumb.jpg", details.Thumbnail)
		assert.Equal(t, "https://youtube.com/watch?v="+expectedID, details.URL)

		expectedPublishedAt, _ := time.Parse(time.RFC3339, publishedDate)
		assert.Equal(t, expectedPublishedAt, details.PublishedAt)
	})

	t.Run("debe retornar error cuando el video ID es inválido", func(t *testing.T) {
		mockLogger := new(logger.MockLogger)
		mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

		client := NewYouTubeClient("test-key", mockLogger)

		_, err := client.GetVideoDetails(context.Background(), "invalid")

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid video ID")
	})

	t.Run("debe retornar error cuando la API devuelve status no OK", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			response := map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Video not found",
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer ts.Close()

		mockLogger := new(logger.MockLogger)
		mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()
		client := NewYouTubeClient("test-key", mockLogger)
		client.BaseURL = ts.URL

		_, err := client.GetVideoDetails(context.Background(), "dQw4w9WgXcQ")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "API de YouTube respondió con código 404")
	})

	t.Run("debe retornar error cuando no hay items en la respuesta", func(t *testing.T) {
		// Arrange
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}}); err != nil {
				t.Fatal(err)
			}
		}))
		defer ts.Close()

		mockLogger := new(logger.MockLogger)
		mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
		mockLogger.On("Warn", mock.Anything, mock.Anything).Return()
		client := NewYouTubeClient("test-key", mockLogger)
		client.BaseURL = ts.URL

		// Act
		_, err := client.GetVideoDetails(context.Background(), "dQw4w9WgXcQ")

		// Assert
		require.Error(t, err)
		assert.EqualError(t, err, "no se encontró el video con el ID proporcionado")
	})

	t.Run("debe retornar error cuando hay un error en la solicitud HTTP", func(t *testing.T) {
		// Arrange
		mockLogger := new(logger.MockLogger)
		mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()

		client := NewYouTubeClient("test-key", mockLogger)
		client.BaseURL = "http://invalid-url-that-doesnt-exist-123456789.com"

		// Act
		_, err := client.GetVideoDetails(context.Background(), "dQw4w9WgXcQ")

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error al hacer la solicitud a la API de YouTube")
	})

	t.Run("debe retornar error cuando la respuesta de la API no se puede decodificar", func(t *testing.T) {
		// Arrange
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{invalid json"))
		}))
		defer ts.Close()

		mockLogger := new(logger.MockLogger)
		mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()

		client := NewYouTubeClient("test-key", mockLogger)
		client.BaseURL = ts.URL

		// Act
		_, err := client.GetVideoDetails(context.Background(), "dQw4w9WgXcQ")

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error al decodificar la respuesta de la API de YouTube")
	})
	t.Run("debe manejar correctamente errores detallados de la API", func(t *testing.T) {
		// Arrange
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			errorResponse := map[string]interface{}{
				"error": map[string]interface{}{
					"message": "API key inválida",
					"errors": []interface{}{
						map[string]interface{}{
							"message": "API key no autorizada",
							"domain":  "youtube.header",
							"reason":  "invalidApiKey",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(errorResponse)
		}))
		defer ts.Close()

		mockLogger := new(logger.MockLogger)
		mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()

		client := NewYouTubeClient("invalid-key", mockLogger)
		client.BaseURL = ts.URL

		// Act
		_, err := client.GetVideoDetails(context.Background(), "dQw4w9WgXcQ")

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "API de YouTube respondió con código 400")
		assert.Contains(t, err.Error(), "API key inválida")
	})
}

func TestYouTubeClient_SearchVideoID(t *testing.T) {
	t.Run("debe retornar videoID cuando la búsqueda es exitosa", func(t *testing.T) {
		// Arrange
		expectedID := "test123"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"id": map[string]interface{}{"videoId": expectedID},
					},
				},
			}); err != nil {
				t.Fatal(err)
			}
		}))
		defer ts.Close()

		mockLogger := new(logger.MockLogger)
		client := NewYouTubeClient("test-key", mockLogger)
		client.BaseURL = ts.URL

		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

		// Act
		videoID, err := client.SearchVideoID(context.Background(), "test query")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedID, videoID)
	})

	t.Run("debe extraer videoID directamente de URL válida", func(t *testing.T) {
		// Arrange
		testURL := "https://youtube.com/watch?v=dQw4w9WgXcQ"
		mockLogger := new(logger.MockLogger)
		client := NewYouTubeClient("test-key", mockLogger)

		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

		// Act
		videoID, err := client.SearchVideoID(context.Background(), testURL)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "dQw4w9WgXcQ", videoID)
	})

	t.Run("debe retornar error cuando la respuesta está vacía", func(t *testing.T) {
		// Arrange
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			if err := json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}}); err != nil {
				t.Fatal(err)
			}
		}))
		defer ts.Close()

		mockLogger := new(logger.MockLogger)
		client := NewYouTubeClient("test-key", mockLogger)
		client.BaseURL = ts.URL

		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
		mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

		// Act
		_, err := client.SearchVideoID(context.Background(), "empty query")

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no se encontraron videos para la consulta")
	})

	t.Run("debe manejar error de la API con detalles", func(t *testing.T) {
		// Arrange
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			errorResponse := map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Parámetro de búsqueda inválido",
					"errors": []interface{}{
						map[string]interface{}{
							"message":  "Consulta demasiado larga",
							"domain":   "youtube.search",
							"reason":   "invalidSearchFilter",
							"location": "q",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(errorResponse)
		}))
		defer ts.Close()

		mockLogger := new(logger.MockLogger)
		client := NewYouTubeClient("test-key", mockLogger)
		client.BaseURL = ts.URL

		mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()

		// Act
		_, err := client.SearchVideoID(context.Background(), "very long query")

		// Assert
		require.Error(t, err)
		var errYouTube *errorsApp.AppError
		assert.True(t, errors.As(err, &errYouTube))
		assert.Contains(t, err.Error(), "domain: youtube.search, reason: invalidSearchFilter")
	})

	t.Run("debe manejar errores de red durante la búsqueda", func(t *testing.T) {
		// Arrange
		mockLogger := new(logger.MockLogger)
		client := NewYouTubeClient("test-key", mockLogger)
		client.BaseURL = "http://invalid-url-that-doesnt-exist-123456789.com"

		mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()

		// Act
		_, err := client.SearchVideoID(context.Background(), "test query")

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error al hacer la solicitud a la API de YouTube")
	})
}

func TestExtractVideoIDFromURL(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected string
		hasError bool
	}{
		{
			name:     "URL estándar",
			url:      "https://youtube.com/watch?v=dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
			hasError: false,
		},
		{
			name:     "URL acortada",
			url:      "https://youtu.be/dQw4w9WgXcQ",
			expected: "dQw4w9WgXcQ",
			hasError: false,
		},
		{
			name:     "URL con parámetros adicionales",
			url:      "https://youtube.com/watch?v=dQw4w9WgXcQ&feature=share",
			expected: "dQw4w9WgXcQ",
			hasError: false,
		},
		{
			name:     "URL inválida",
			url:      "https://invalid.com/video",
			expected: "",
			hasError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result, err := ExtractVideoIDFromURL(tc.url)

			// Assert
			if tc.hasError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "URL de YouTube invalida")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}
