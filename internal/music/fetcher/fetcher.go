package fetcher

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/internal/music"
	"io"
)

// Fetcher define un contrato para obtener metadatos y datos de audio de servicios de música.
type Fetcher interface {
	LookupSongs(ctx context.Context, query string) ([]*music.Song, error)
	GetDCAData(ctx context.Context, song *music.Song) (io.Reader, error)
}
