package store

import "github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"

// StateStorage define métodos para el almacenamiento y manipulación del estado del reproductor de música.
type StateStorage interface {
	// GetCurrentSong devuelve la canción que se está reproduciendo actualmente.
	GetCurrentSong() (*voice.PlayedSong, error)
	// SetCurrentSong establece la canción que se está reproduciendo actualmente.
	SetCurrentSong(*voice.PlayedSong) error
	// GetVoiceChannel devuelve el ID del canal de voz actual.
	GetVoiceChannel() (string, error)
	// SetVoiceChannel establece el ID del canal de voz actual.
	SetVoiceChannel(string) error
	// GetTextChannel devuelve el ID del canal de texto actual.
	GetTextChannel() (string, error)
	// SetTextChannel establece el ID del canal de texto actual.
	SetTextChannel(string) error
}
