package ports

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

// SongStorage define métodos para el almacenamiento y manipulación de la lista de reproducción de canciones.
type SongStorage interface {
	// PrependSong agrega una canción al principio de la lista de reproducción.
	PrependSong(*entity.Song) error
	// AppendSong agrega una canción al final de la lista de reproducción.
	AppendSong(*entity.Song) error
	// RemoveSong elimina una canción de la lista de reproducción por su posición.
	RemoveSong(int) (*entity.Song, error)
	// ClearPlaylist elimina todas las canciones de la lista de reproducción.
	ClearPlaylist() error
	// GetSongs devuelve todas las canciones en la lista de reproducción.
	GetSongs() ([]*entity.Song, error)
	// PopFirstSong elimina y devuelve la primera canción de la lista de reproducción.
	PopFirstSong() (*entity.Song, error)
}
