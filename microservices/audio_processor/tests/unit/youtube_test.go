package unit

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestYouTubeClient(t *testing.T) {
	t.Run("GetVideoDetails", func(t *testing.T) {
		t.Run("Successful request", func(t *testing.T) {
			// Configuración de un servidor de pruebas para simular respuestas de la API de YouTube
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := `{
					"items": [{
						"id": "test-video-id",
						"snippet": {
							"title": "Test Video",
							"URLYouTube": "https://youtube.com/watch?v=test-video-id",
							"description": "This is a test video",
							"channelTitle": "Test Channel",
							"publishedAt": "2022-01-01T00:00:00Z",
							"thumbnails": {
								"default": {
									"url": "test-url.jpg"
								}
							}
						},
						"contentDetails": {
							"duration": "PT1H2M3S"
						}
					}]
				}`
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))
			}))
			defer server.Close()

			client := &api.YouTubeClient{
				ApiKey:     "test-api-key",
				BaseURL:    server.URL,
				HttpClient: server.Client(),
			}

			ctx := context.Background()
			details, err := client.GetVideoDetails(ctx, "test-video-id")

			if err != nil {
				t.Fatalf("Error inesperado: %v", err)
			}

			expectedPublishedAt, _ := time.Parse(time.RFC3339, "2022-01-01T00:00:00Z")
			expected := &api.VideoDetails{
				Title:       "Test Video",
				VideoID:     "test-video-id",
				Description: "This is a test video",
				ChannelName: "Test Channel",
				Duration:    "PT1H2M3S",
				PublishedAt: expectedPublishedAt,
				URLYouTube:  "https://youtube.com/watch?v=test-video-id",
				Thumbnail:   "test-url.jpg",
			}

			if *details != *expected {
				t.Errorf("Se esperaba %+v, pero se obtuvo %+v", expected, details)
			}
		})

		t.Run("API error", func(t *testing.T) {
			// Configuración de un servidor de pruebas para simular un error en la API de YouTube
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			}))
			defer server.Close()

			client := &api.YouTubeClient{
				ApiKey:     "test-api-key",
				BaseURL:    server.URL,
				HttpClient: server.Client(),
			}

			ctx := context.Background()
			_, err := client.GetVideoDetails(ctx, "test-video-id")

			if err == nil {
				t.Fatal("Se esperaba un error, pero se obtuvo nil")
			}
		})
	})

	t.Run("SearchVideoID", func(t *testing.T) {
		t.Run("Successful request", func(t *testing.T) {
			// Configuración de un servidor de pruebas para simular una búsqueda exitosa en la API de YouTube
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := `{
					"items": [{
						"id": {
							"videoId": "test-video-id"
						}
					}]
				}`
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))
			}))
			defer server.Close()

			client := &api.YouTubeClient{
				ApiKey:     "test-api-key",
				BaseURL:    server.URL,
				HttpClient: server.Client(),
			}

			ctx := context.Background()
			videoID, err := client.SearchVideoID(ctx, "test query")

			if err != nil {
				t.Fatalf("Error inesperado: %v", err)
			}

			expected := "test-video-id"
			if videoID != expected {
				t.Errorf("Se esperaba %s, pero se obtuvo %s", expected, videoID)
			}
		})

		t.Run("No results", func(t *testing.T) {
			// Configuración de un servidor de pruebas para simular una búsqueda sin resultados en la API de YouTube
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := `{"items": []}`
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))
			}))
			defer server.Close()

			client := &api.YouTubeClient{
				ApiKey:     "test-api-key",
				BaseURL:    server.URL,
				HttpClient: server.Client(),
			}

			ctx := context.Background()
			_, err := client.SearchVideoID(ctx, "test query")

			if err == nil {
				t.Fatal("Se esperaba un error, pero se obtuvo nil")
			}
		})
	})
}
