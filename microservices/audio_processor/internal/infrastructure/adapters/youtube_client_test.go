//go:build !integration

package adapters

import (
	"context"
	"encoding/json"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
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
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"id": expectedID,
						"snippet": map[string]interface{}{
							"title":        "Test Video",
							"description":  "Test Description",
							"channelTitle": "Test Channel",
							"publishedAt":  time.Now().Format(time.RFC3339),
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
			json.NewEncoder(w).Encode(response)
		}))
		defer ts.Close()

		log, err := logger.NewZapLogger()
		require.NoError(t, err)
		client := NewYouTubeClient("test-key", log)
		client.BaseURL = ts.URL

		// Act
		details, err := client.GetVideoDetails(context.Background(), expectedID)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedID, details.ID)
		assert.Equal(t, "Test Video", details.Title)
		assert.Equal(t, "PT5M30S", details.Duration)
	})

	t.Run("debe retornar error cuando la API devuelve status no OK", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer ts.Close()

		log, err := logger.NewZapLogger()
		require.NoError(t, err)
		client := NewYouTubeClient("test-key", log)
		client.BaseURL = ts.URL

		_, err = client.GetVideoDetails(context.Background(), "invalidID")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "API respondió con código 404")
	})

	t.Run("debe retornar error cuando no hay items en la respuesta", func(t *testing.T) {
		// Arrange
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}})
		}))
		defer ts.Close()

		log, err := logger.NewZapLogger()
		require.NoError(t, err)
		client := NewYouTubeClient("test-key", log)
		client.BaseURL = ts.URL

		// Act
		_, err = client.GetVideoDetails(context.Background(), "emptyID")

		// Assert
		require.Error(t, err)
		assert.EqualError(t, err, "no se encontró el video con el ID proporcionado")
	})
}

func TestYouTubeClient_SearchVideoID(t *testing.T) {
	t.Run("debe retornar videoID cuando la búsqueda es exitosa", func(t *testing.T) {
		// Arrange
		expectedID := "test123"
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{
						"id": map[string]interface{}{"videoId": expectedID},
					},
				},
			})
		}))
		defer ts.Close()

		log, err := logger.NewZapLogger()
		client := NewYouTubeClient("test-key", log)
		client.BaseURL = ts.URL

		// Act
		videoID, err := client.SearchVideoID(context.Background(), "test query")

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedID, videoID)
	})

	t.Run("debe extraer videoID directamente de URL válida", func(t *testing.T) {
		// Arrange
		testURL := "https://youtube.com/watch?v=dQw4w9WgXcQ"
		log, err := logger.NewZapLogger()
		client := NewYouTubeClient("test-key", log)

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
			json.NewEncoder(w).Encode(map[string]interface{}{"items": []interface{}{}})
		}))
		defer ts.Close()

		log, err := logger.NewZapLogger()
		require.NoError(t, err)
		client := NewYouTubeClient("test-key", log)
		client.BaseURL = ts.URL

		// Act
		_, err = client.SearchVideoID(context.Background(), "empty query")

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no se encontraron videos para la consulta")
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
