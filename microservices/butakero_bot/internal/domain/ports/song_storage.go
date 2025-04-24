package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

// PlaylistStorage define métodos para el almacenamiento y manipulación de la lista de reproducción de canciones.
type PlaylistStorage interface {
	// PrependTrack agrega una canción al principio de la lista de reproducción.
	PrependTrack(ctx context.Context, track *entity.PlayedSong) error
	// AppendTrack agrega una canción al final de la lista de reproducción.
	AppendTrack(ctx context.Context, track *entity.PlayedSong) error
	// RemoveTrack elimina una canción de la lista de reproducción por su posición.
	RemoveTrack(ctx context.Context, position int) (*entity.PlayedSong, error)
	// ClearPlaylist elimina todas las canciones de la lista de reproducción.
	ClearPlaylist(ctx context.Context) error
	// GetAllTracks devuelve todas las canciones en la lista de reproducción.
	GetAllTracks(ctx context.Context) ([]*entity.PlayedSong, error)
	// PopNextTrack elimina y devuelve la primera canción de la lista de reproducción.
	PopNextTrack(ctx context.Context) (*entity.PlayedSong, error)
}
