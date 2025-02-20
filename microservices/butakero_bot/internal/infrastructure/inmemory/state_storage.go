package inmemory

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"sync"
)

// InmemoryStateStorage implementa la interfaz StateStorage utilizando la memoria RAM para almacenar el estado del reproductor.
type InmemoryStateStorage struct {
	mutex        sync.RWMutex
	currentSong  *entity.PlayedSong
	textChannel  string
	voiceChannel string
	logger       logging.Logger
}

// NewInmemoryStateStorage crea una nueva instancia de InmemoryStateStorage.
func NewInmemoryStateStorage(logger logging.Logger) *InmemoryStateStorage {
	return &InmemoryStateStorage{
		mutex:  sync.RWMutex{},
		logger: logger,
	}
}

// GetCurrentSong devuelve la canción actual que se está reproduciendo.
func (s *InmemoryStateStorage) GetCurrentSong() (*entity.PlayedSong, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.currentSong, nil
}

// SetCurrentSong establece la canción actual que se está reproduciendo.
func (s *InmemoryStateStorage) SetCurrentSong(song *entity.PlayedSong) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.currentSong = song
	return nil
}

// GetVoiceChannel devuelve el ID del canal de voz.
func (s *InmemoryStateStorage) GetVoiceChannel() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.logger.Info("Obteniendo el canal de voz")
	return s.voiceChannel, nil
}

// SetVoiceChannel establece el ID del canal de voz.
func (s *InmemoryStateStorage) SetVoiceChannel(channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.voiceChannel = channelID
	s.logger.Info("Canal de voz establecido")
	return nil
}

// GetTextChannel devuelve el ID del canal de texto.
func (s *InmemoryStateStorage) GetTextChannel() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.logger.Info("Obteniendo el canal de texto")
	return s.textChannel, nil
}

// SetTextChannel establece el ID del canal de texto.
func (s *InmemoryStateStorage) SetTextChannel(channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.textChannel = channelID
	s.logger.Info("Canal de texto establecido")
	return nil
}
