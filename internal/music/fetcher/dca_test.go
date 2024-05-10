package fetcher

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"io"
	"testing"
)

func TestGetDCAData(t *testing.T) {
	ctx := context.Background()

	// Caso 1: Tipo de canción compatible (yt-dlp)
	t.Run("SupportedSongType", func(t *testing.T) {
		song := &bot.Song{
			Type: "yt-dlp",
			URL:  "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		}

		reader, err := GetDCAData(ctx, song)

		if err != nil {
			t.Errorf("GetDCAData devolvió un error inesperado para el tipo de canción compatible: %v", err)
		}

		data := make([]byte, 1024)
		n, err := reader.Read(data)
		if err != nil && err != io.EOF {
			t.Errorf("Error al leer desde el lector de datos DCA para el tipo de canción compatible: %v", err)
		}
		if n == 0 {
			t.Error("GetDCAData no devolvió ningún dato DCA para el tipo de canción compatible")
		}
	})

	// Caso 2: Tipo de canción no compatible
	t.Run("UnsupportedSongType", func(t *testing.T) {
		song := &bot.Song{
			Type: "unsupported-type",
			URL:  "https://example.com/unsupported-song",
		}

		_, err := GetDCAData(ctx, song)

		if err == nil {
			t.Error("GetDCAData no devolvió un error para el tipo de canción no compatible")
		}
		expectedError := fmt.Sprintf("tipo de musica no soportada: %s", song.Type)
		if err.Error() != expectedError {
			t.Errorf("GetDCAData devolvió un mensaje de error inesperado para el tipo de canción no compatible. Recibido: %v, Esperado: %v", err.Error(), expectedError)
		}
	})
}
