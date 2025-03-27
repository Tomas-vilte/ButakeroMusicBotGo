package inmemory

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
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

	logger := s.logger.With(zap.String("method", "GetCurrentSong"))
	logger.Debug("Obteniendo la canción actual")
	return s.currentSong, nil
}

// SetCurrentSong establece la canción actual que se está reproduciendo.
func (s *InmemoryStateStorage) SetCurrentSong(song *entity.PlayedSong) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := s.logger.With(
		zap.String("method", "SetCurrentSong"),
	)

	if song != nil {
		logger = logger.With(zap.String("songTitle", song.DiscordSong.TitleTrack))
	}

	s.currentSong = song
	if song == nil {
		logger.Info("Canción actual limpiada")
	} else {
		logger.Info("Canción actual establecida")
	}
	return nil
}

// GetVoiceChannel devuelve el ID del canal de voz.
func (s *InmemoryStateStorage) GetVoiceChannel() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	logger := s.logger.With(zap.String("method", "GetVoiceChannel"))
	logger.Debug("Obteniendo el canal de voz")
	return s.voiceChannel, nil
}

// SetVoiceChannel establece el ID del canal de voz.
func (s *InmemoryStateStorage) SetVoiceChannel(channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := s.logger.With(
		zap.String("method", "SetVoiceChannel"),
		zap.String("voiceChannelID", channelID),
	)

	s.voiceChannel = channelID
	logger.Info("Canal de voz establecido")
	return nil
}

// GetTextChannel devuelve el ID del canal de texto.
func (s *InmemoryStateStorage) GetTextChannel() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	logger := s.logger.With(zap.String("method", "GetTextChannel"))
	logger.Debug("Obteniendo el canal de texto")
	return s.textChannel, nil
}

// SetTextChannel establece el ID del canal de texto.
func (s *InmemoryStateStorage) SetTextChannel(channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := s.logger.With(
		zap.String("method", "SetTextChannel"),
		zap.String("channelID", channelID),
	)

	s.textChannel = channelID
	logger.Info("Canal de texto establecido")
	return nil
}
