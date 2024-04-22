package fetcher

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"io"
)

// GetDCAData obtiene datos de audio DCA para una canción.
func GetDCAData(ctx context.Context, song *bot.Song) (io.Reader, error) {
	switch song.Type {
	case "yt-dlp":
		return NewYoutubeFetcher().GetDCAData(ctx, song)
	}
	return nil, fmt.Errorf("tipo de musica no soportada: %s", song.Type)
}
