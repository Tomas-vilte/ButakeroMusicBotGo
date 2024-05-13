package discord

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
)

// InMemoryInteractionStorage es una estructura de almacenamiento en memoria para interacciones.
type InMemoryInteractionStorage struct {
	songsToAdd map[string][]*voice.Song
}

// NewInMemoryStorage crea una nueva instancia de InMemoryInteractionStorage con un mapa inicializado.
func NewInMemoryStorage() *InMemoryInteractionStorage {
	return &InMemoryInteractionStorage{
		songsToAdd: make(map[string][]*voice.Song),
	}
}

// SaveSongList guarda una lista de canciones en el canal identificado por channelID.
func (s *InMemoryInteractionStorage) SaveSongList(channelID string, list []*voice.Song) {
	s.songsToAdd[channelID] = list
}

// DeleteSongList elimina la lista de canciones asociada al canal identificado por channelID.
func (s *InMemoryInteractionStorage) DeleteSongList(channelID string) {
	delete(s.songsToAdd, channelID)
}

// GetSongList retorna la lista de canciones asociada al canal identificado por channelID.
func (s *InMemoryInteractionStorage) GetSongList(channelID string) []*voice.Song {
	return s.songsToAdd[channelID]
}
