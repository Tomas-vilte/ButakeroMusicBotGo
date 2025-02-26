package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

// SongDownloader define métodos para la descarga de canciones.
// Inicia la descarga de una canción si no existe, llamando a un microservicio de descarga.
type SongDownloader interface {
	// DownloadSong descarga una canción dado su nombre.
	DownloadSong(ctx context.Context, songName string) (*entity.DownloadResponse, error)
}
