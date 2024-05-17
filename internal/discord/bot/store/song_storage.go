package store

import "github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"

// SongStorage define métodos para el almacenamiento y manipulación de la lista de reproducción de canciones.
type SongStorage interface {
	// PrependSong agrega una canción al principio de la lista de reproducción.
	PrependSong(*voice.Song) error
	// AppendSong agrega una canción al final de la lista de reproducción.
	AppendSong(*voice.Song) error
	// RemoveSong elimina una canción de la lista de reproducción por su posición.
	RemoveSong(int) (*voice.Song, error)
	// ClearPlaylist elimina todas las canciones de la lista de reproducción.
	ClearPlaylist() error
	// GetSongs devuelve todas las canciones en la lista de reproducción.
	GetSongs() ([]*voice.Song, error)
	// PopFirstSong elimina y devuelve la primera canción de la lista de reproducción.
	PopFirstSong() (*voice.Song, error)
}
