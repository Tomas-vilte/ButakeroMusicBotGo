package integration

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
	"os"
	"testing"
)

// TestYouTubeClientIntegration contiene pruebas de integración que requieren una API key real.
// Estas pruebas están deshabilitadas por defecto para evitar llamadas no deseadas a la API.
func TestYouTubeClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Saltando test de integración en modo corto")
	}

	apiKey := os.Getenv("YOUTUBE_API_KEY")
	client := api.NewYouTubeClient(apiKey)

	t.Run("GetVideoDetails Integration", func(t *testing.T) {
		ctx := context.Background()
		videoID := "dQw4w9WgXcQ" // ID del video "Never Gonna Give You Up" de Rick Astley
		details, err := client.GetVideoDetails(ctx, videoID)

		if err != nil {
			t.Fatalf("Error al obtener detalles del video: %v", err)
		}

		if details.Title == "" {
			t.Error("Se esperaba un título no vacío")
		}
		if details.ChannelName == "" {
			t.Error("Se esperaba un nombre de canal no vacío")
		}
		if details.Duration == "" {
			t.Error("Se esperaba una duración no vacía")
		}
	})

	t.Run("SearchVideoID by URL", func(t *testing.T) {
		ctx := context.Background()
		videoID, err := client.SearchVideoID(ctx, "https://www.youtube.com/watch?v=dQw4w9WgXcQ")

		if err != nil {
			t.Fatalf("Error al buscar video por URL: %v", err)
		}

		if videoID == "" {
			t.Error("Se esperaba un ID de video no vacío")
		}
	})

	t.Run("SearchVideoID by Term", func(t *testing.T) {
		ctx := context.Background()
		videoID, err := client.SearchVideoID(ctx, "Never Gonna Give You Up")

		if err != nil {
			t.Fatalf("Error al buscar video por término: %v", err)
		}

		if videoID == "" {
			t.Error("Se esperaba un ID de video no vacío")
		}
	})
}
