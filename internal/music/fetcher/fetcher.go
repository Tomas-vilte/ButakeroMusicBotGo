package fetcher

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"io"
)

// Fetcher define un contrato para obtener metadatos y datos de audio de servicios de música.
type Fetcher interface {
	LookupSongs(ctx context.Context, query string) ([]*bot.Song, error)
	GetDCAData(ctx context.Context, song *bot.Song) (io.Reader, error)
}
