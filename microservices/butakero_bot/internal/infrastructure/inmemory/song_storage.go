package inmemory

import (
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
	"sync"
)

var (
	ErrNoSongs               = errors.New("canción no disponible")
	ErrRemoveInvalidPosition = errors.New("posición inválida")
)

// InmemorySongStorage implementa la interfaz SongStorage utilizando la memoria ram para almacenar la lista de reproducción de canciones.
type InmemorySongStorage struct {
	mutex  sync.RWMutex
	songs  []*entity.PlayedSong
	logger logging.Logger
}

// NewInmemorySongStorage crea una nueva instancia de InmemorySongStorage.
func NewInmemorySongStorage(logger logging.Logger) *InmemorySongStorage {
	return &InmemorySongStorage{
		mutex:  sync.RWMutex{},
		songs:  make([]*entity.PlayedSong, 0),
		logger: logger,
	}
}

// PrependSong agrega una canción al principio de la lista de reproducción.
func (s *InmemorySongStorage) PrependSong(song *entity.PlayedSong) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := s.logger.With(
		zap.String("method", "PrependSong"),
		zap.String("songTitle", song.DiscordSong.TitleTrack),
	)

	s.songs = append([]*entity.PlayedSong{song}, s.songs...)
	logger.Info("Canción agregada al principio de la lista de reproducción")
	return nil
}

// AppendSong agrega una canción al final de la lista de reproducción.
func (s *InmemorySongStorage) AppendSong(song *entity.PlayedSong) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := s.logger.With(
		zap.String("method", "AppendSong"),
		zap.String("songTitle", song.DiscordSong.TitleTrack),
	)

	s.songs = append(s.songs, song)
	logger.Info("Canción agregada al final de la lista de reproducción")
	return nil
}

// RemoveSong elimina una canción de la lista de reproducción por posición.
func (s *InmemorySongStorage) RemoveSong(position int) (*entity.PlayedSong, error) {
	index := position - 1

	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := s.logger.With(
		zap.String("method", "RemoveSong"),
		zap.Int("position", position),
	)

	if index >= len(s.songs) || index < 0 {
		logger.Info("Posición de eliminación de canción inválida")
		return nil, ErrRemoveInvalidPosition
	}

	song := s.songs[index]

	copy(s.songs[index:], s.songs[index+1:])
	s.songs[len(s.songs)-1] = nil
	s.songs = s.songs[:len(s.songs)-1]
	logger.Info("Canción eliminada de la lista de reproducción")
	return song, nil
}

// ClearPlaylist elimina todas las canciones de la lista de reproducción.
func (s *InmemorySongStorage) ClearPlaylist() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := s.logger.With(zap.String("method", "ClearPlaylist"))
	logger.Info("Lista de reproducción borrada")

	s.songs = make([]*entity.PlayedSong, 0)
	return nil
}

// GetSongs devuelve todas las canciones de la lista de reproducción.
func (s *InmemorySongStorage) GetSongs() ([]*entity.PlayedSong, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	logger := s.logger.With(zap.String("method", "GetSongs"))
	logger.Info("Obteniendo todas las canciones de la lista de reproducción")

	// Se copian las canciones para evitar modificaciones inadvertidas.
	songs := make([]*entity.PlayedSong, len(s.songs))
	copy(songs, s.songs)

	return songs, nil
}

// PopFirstSong elimina y devuelve la primera canción de la lista de reproducción.
func (s *InmemorySongStorage) PopFirstSong() (*entity.PlayedSong, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := s.logger.With(zap.String("method", "PopFirstSong"))

	if len(s.songs) == 0 {
		logger.Info("No hay canciones para eliminar")
		return nil, ErrNoSongs
	}

	song := s.songs[0]
	s.songs = s.songs[1:]
	logger.Info("Primera canción eliminada de la lista de reproducción")
	return song, nil
}
