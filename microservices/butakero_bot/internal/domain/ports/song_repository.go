package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

// SongRepository define los métodos que cualquier implementación de un repositorio de canciones debe tener.
type SongRepository interface {
	// GetSongByVideoID obtiene una canción por su ID de video.
	GetSongByVideoID(ctx context.Context, videoID string) (*entity.SongEntity, error)
	// SearchSongsByTitle busca canciones por su título.
	SearchSongsByTitle(ctx context.Context, title string) ([]*entity.SongEntity, error)
}
