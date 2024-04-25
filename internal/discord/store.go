package discord

import "github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"

// InMemoryInteractionStorage es una estructura de almacenamiento en memoria para interacciones.
type InMemoryInteractionStorage struct {
	songsToAdd map[string][]*bot.Song
}

// NewInMemoryStorage crea una nueva instancia de InMemoryInteractionStorage con un mapa inicializado.
func NewInMemoryStorage() *InMemoryInteractionStorage {
	return &InMemoryInteractionStorage{
		songsToAdd: make(map[string][]*bot.Song),
	}
}

// SaveSongList guarda una lista de canciones en el canal identificado por channelID.
func (s *InMemoryInteractionStorage) SaveSongList(channelID string, list []*bot.Song) {
	s.songsToAdd[channelID] = list
}

// DeleteSongList elimina la lista de canciones asociada al canal identificado por channelID.
func (s *InMemoryInteractionStorage) DeleteSongList(channelID string) {
	delete(s.songsToAdd, channelID)
}

// GetSongList retorna la lista de canciones asociada al canal identificado por channelID.
func (s *InMemoryInteractionStorage) GetSongList(channelID string) []*bot.Song {
	return s.songsToAdd[channelID]
}
