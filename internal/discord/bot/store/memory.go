package store

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"go.uber.org/zap"
	"sync"
)

// InmemoryPlaylistStorage es una implementación de GuildPlayerState que almacena la lista de reproducción en memoria.
type InmemoryPlaylistStorage struct {
	mutex        sync.RWMutex
	songs        []*bot.Song     // Lista de canciones.
	currentSong  *bot.PlayedSong // Canción actual que se está reproduciendo.
	textChannel  string          // ID del canal de texto.
	voiceChannel string          // ID del canal de voz.
	logger       *zap.Logger
}

// NewInmemoryGuildPlayerState crea una nueva instancia de InmemoryPlaylistStorage.
func NewInmemoryGuildPlayerState(logger *zap.Logger) *InmemoryPlaylistStorage {
	return &InmemoryPlaylistStorage{
		mutex:  sync.RWMutex{},
		songs:  make([]*bot.Song, 0),
		logger: logger,
	}
}

// GetCurrentSong devuelve la canción actual que se está reproduciendo.
func (s *InmemoryPlaylistStorage) GetCurrentSong() (*bot.PlayedSong, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.logger.Debug("Obteniendo la canción actual")
	return s.currentSong, nil
}

// SetCurrentSong establece la canción actual que se está reproduciendo.
func (s *InmemoryPlaylistStorage) SetCurrentSong(song *bot.PlayedSong) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.currentSong = song
	s.logger.Debug("Canción actual establecida")
	return nil
}

// GetVoiceChannel devuelve el ID del canal de voz.
func (s *InmemoryPlaylistStorage) GetVoiceChannel() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.logger.Debug("Obteniendo el canal de voz")
	return s.voiceChannel, nil
}

// SetVoiceChannel establece el ID del canal de voz.
func (s *InmemoryPlaylistStorage) SetVoiceChannel(channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.voiceChannel = channelID
	s.logger.Debug("Canal de voz establecido")
	return nil
}

// GetTextChannel devuelve el ID del canal de texto.
func (s *InmemoryPlaylistStorage) GetTextChannel() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.logger.Debug("Obteniendo el canal de texto")
	return s.textChannel, nil
}

// SetTextChannel establece el ID del canal de texto.
func (s *InmemoryPlaylistStorage) SetTextChannel(channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.textChannel = channelID
	s.logger.Debug("Canal de texto establecido")
	return nil
}

// PrependSong agrega una canción al principio de la lista de reproducción.
func (s *InmemoryPlaylistStorage) PrependSong(song *bot.Song) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.songs = append([]*bot.Song{song}, s.songs...)
	s.logger.Debug("Canción agregada al principio de la lista de reproducción")
	return nil
}

// AppendSong agrega una canción al final de la lista de reproducción.
func (s *InmemoryPlaylistStorage) AppendSong(song *bot.Song) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.songs = append(s.songs, song)
	s.logger.Debug("Canción agregada al final de la lista de reproducción")
	return nil
}

// RemoveSong elimina una canción de la lista de reproducción por posición.
func (s *InmemoryPlaylistStorage) RemoveSong(position int) (*bot.Song, error) {
	index := position - 1

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if index >= len(s.songs) || index < 0 {
		s.logger.Error("Posición de eliminación de canción inválida")
		return nil, bot.ErrRemoveInvalidPosition
	}

	song := s.songs[index]

	copy(s.songs[index:], s.songs[index+1:])
	s.songs[len(s.songs)-1] = nil
	s.songs = s.songs[:len(s.songs)-1]
	s.logger.Debug("Canción eliminada de la lista de reproducción")
	return song, nil
}

// ClearPlaylist elimina todas las canciones de la lista de reproducción.
func (s *InmemoryPlaylistStorage) ClearPlaylist() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.songs = make([]*bot.Song, 0)
	s.logger.Debug("Lista de reproducción borrada")
	return nil
}

// GetSongs devuelve todas las canciones de la lista de reproducción.
func (s *InmemoryPlaylistStorage) GetSongs() ([]*bot.Song, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Se copian las canciones para evitar modificaciones inadvertidas.
	songs := make([]*bot.Song, len(s.songs))
	copy(songs, s.songs)

	s.logger.Debug("Obteniendo todas las canciones de la lista de reproducción")
	return songs, nil
}

// PopFirstSong elimina y devuelve la primera canción de la lista de reproducción.
func (s *InmemoryPlaylistStorage) PopFirstSong() (*bot.Song, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.songs) == 0 {
		s.logger.Error("No hay canciones para eliminar")
		return nil, bot.ErrNoSongs
	}

	song := s.songs[0]
	s.songs = s.songs[1:]
	s.logger.Debug("Primera canción eliminada de la lista de reproducción")
	return song, nil
}
