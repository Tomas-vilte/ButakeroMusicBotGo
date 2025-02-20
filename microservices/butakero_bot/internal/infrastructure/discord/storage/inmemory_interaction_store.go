package storage

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
	"sync"
)

// InMemoryInteractionStorage es una estructura de almacenamiento en memoria para interacciones.
type InMemoryInteractionStorage struct {
	songsToAdd map[string][]*entity.Song
	mutex      sync.RWMutex
	logger     logging.Logger
}

func NewInMemoryInteractionStorage(logger logging.Logger) *InMemoryInteractionStorage {
	return &InMemoryInteractionStorage{
		songsToAdd: make(map[string][]*entity.Song),
		logger:     logger,
	}
}

// SaveSongList guarda una lista de canciones en el canal identificado por channelID.
func (s *InMemoryInteractionStorage) SaveSongList(channelID string, list []*entity.Song) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.songsToAdd[channelID] = list
	s.logger.Info("Lista de canciones guardada", zap.String("channelID", channelID), zap.Any("lista", list))
}

// DeleteSongList elimina la lista de canciones asociada al canal identificado por channelID.
func (s *InMemoryInteractionStorage) DeleteSongList(channelID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.songsToAdd, channelID)
	s.logger.Info("Lista de canciones eliminada", zap.String("channelID", channelID))
}

// GetSongList retorna la lista de canciones asociada al canal identificado por channelID.
func (s *InMemoryInteractionStorage) GetSongList(channelID string) []*entity.Song {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	list, exists := s.songsToAdd[channelID]
	if !exists {
		s.logger.Warn("No se encontró la lista de canciones", zap.String("channelID", channelID))
		return nil
	}
	return list
}
