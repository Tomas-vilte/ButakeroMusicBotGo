package inmemory

import (
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"sync"
)

var (
	ErrNoSongs               = errors.New("canción no disponible")
	ErrRemoveInvalidPosition = errors.New("posición inválida")
)

// InmemorySongStorage implementa la interfaz SongStorage utilizando la memoria ram para almacenar la lista de reproducción de canciones.
type InmemorySongStorage struct {
	mutex  sync.RWMutex
	songs  []*entity.Song
	logger logging.Logger
}

// NewInmemorySongStorage crea una nueva instancia de InmemorySongStorage.
func NewInmemorySongStorage(logger logging.Logger) *InmemorySongStorage {
	return &InmemorySongStorage{
		mutex:  sync.RWMutex{},
		songs:  make([]*entity.Song, 0),
		logger: logger,
	}
}

// PrependSong agrega una canción al principio de la lista de reproducción.
func (s *InmemorySongStorage) PrependSong(song *entity.Song) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.songs = append([]*entity.Song{song}, s.songs...)
	s.logger.Info("Canción agregada al principio de la lista de reproducción")
	return nil
}

// AppendSong agrega una canción al final de la lista de reproducción.
func (s *InmemorySongStorage) AppendSong(song *entity.Song) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.songs = append(s.songs, song)
	s.logger.Info("Canción agregada al final de la lista de reproducción")
	return nil
}

// RemoveSong elimina una canción de la lista de reproducción por posición.
func (s *InmemorySongStorage) RemoveSong(position int) (*entity.Song, error) {
	index := position - 1

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if index >= len(s.songs) || index < 0 {
		s.logger.Info("Posición de eliminación de canción inválida")
		return nil, ErrRemoveInvalidPosition
	}

	song := s.songs[index]

	copy(s.songs[index:], s.songs[index+1:])
	s.songs[len(s.songs)-1] = nil
	s.songs = s.songs[:len(s.songs)-1]
	s.logger.Info("Canción eliminada de la lista de reproducción")
	return song, nil
}

// ClearPlaylist elimina todas las canciones de la lista de reproducción.
func (s *InmemorySongStorage) ClearPlaylist() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.songs = make([]*entity.Song, 0)
	s.logger.Info("Lista de reproducción borrada")
	return nil
}

// GetSongs devuelve todas las canciones de la lista de reproducción.
func (s *InmemorySongStorage) GetSongs() ([]*entity.Song, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Se copian las canciones para evitar modificaciones inadvertidas.
	songs := make([]*entity.Song, len(s.songs))
	copy(songs, s.songs)

	s.logger.Info("Obteniendo todas las canciones de la lista de reproducción")
	return songs, nil
}

// PopFirstSong elimina y devuelve la primera canción de la lista de reproducción.
func (s *InmemorySongStorage) PopFirstSong() (*entity.Song, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.songs) == 0 {
		s.logger.Info("No hay canciones para eliminar")
		return nil, ErrNoSongs
	}

	song := s.songs[0]
	s.songs = s.songs[1:]
	s.logger.Info("Primera canción eliminada de la lista de reproducción")
	return song, nil
}
