package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

// SongService es la interfaz que define los métodos para manejar canciones
type SongService interface {
	// GetOrDownloadSong inicia el proceso de descarga de una canción al otro servicio mediante colas
	GetOrDownloadSong(ctx context.Context, userID, songInput, providerType string) (*entity.DiscordEntity, error)
}
