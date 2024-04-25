package discord

import "github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"

type InMemoryInteractionStorage struct {
	songsToAdd map[string][]*bot.Song
}

func NewInMemoryStorage() *InMemoryInteractionStorage {
	return &InMemoryInteractionStorage{
		songsToAdd: make(map[string][]*bot.Song),
	}
}

func (s *InMemoryInteractionStorage) SaveSongList(channelID string, list []*bot.Song) {
	s.songsToAdd[channelID] = list
}

func (s *InMemoryInteractionStorage) DeleteSongList(channelID string) {
	delete(s.songsToAdd, channelID)
}

func (s *InMemoryInteractionStorage) GetSongList(channelID string) []*bot.Song {
	return s.songsToAdd[channelID]
}
