package fetcher

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"io"
)

// Fetcher define un contrato para obtener metadatos y datos de audio de servicios de música.
type Fetcher interface {
	LookupSongs(ctx context.Context, query string) ([]*voice.Song, error)
	GetDCAData(ctx context.Context, song *voice.Song) (io.Reader, error)
}
