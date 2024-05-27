package inmemory_storage

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"sync"
)

// InmemorySongStorage implementa la interfaz SongStorage utilizando la memoria ram para almacenar la lista de reproducción de canciones.
type InmemorySongStorage struct {
	mutex  sync.RWMutex   // mutex se utiliza para garantizar la concurrencia segura al manipular la lista de reproducción.
	songs  []*voice.Song  // songs es la lista de reproducción de canciones almacenada en memoria.
	logger logging.Logger // logger es un registrador para registrar mensajes de depuración y errores.
}

// NewInmemorySongStorage crea una nueva instancia de InmemorySongStorage.
func NewInmemorySongStorage(logger logging.Logger) *InmemorySongStorage {
	return &InmemorySongStorage{
		mutex:  sync.RWMutex{},         // Se inicializa un nuevo mutex para garantizar la concurrencia segura.
		songs:  make([]*voice.Song, 0), // Se inicializa una nueva lista de reproducción de canciones vacía.
		logger: logger,                 // Se inicializa un nuevo logger con un logger "Nop" (sin operación) por defecto.
	}
}

// PrependSong agrega una canción al principio de la lista de reproducción.
func (s *InmemorySongStorage) PrependSong(song *voice.Song) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.songs = append([]*voice.Song{song}, s.songs...)
	s.logger.Info("Canción agregada al principio de la lista de reproducción")
	return nil
}

// AppendSong agrega una canción al final de la lista de reproducción.
func (s *InmemorySongStorage) AppendSong(song *voice.Song) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.songs = append(s.songs, song)
	s.logger.Info("Canción agregada al final de la lista de reproducción")
	return nil
}

// RemoveSong elimina una canción de la lista de reproducción por posición.
func (s *InmemorySongStorage) RemoveSong(position int) (*voice.Song, error) {
	index := position - 1

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if index >= len(s.songs) || index < 0 {
		s.logger.Info("Posición de eliminación de canción inválida")
		return nil, bot.ErrRemoveInvalidPosition
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

	s.songs = make([]*voice.Song, 0)
	s.logger.Info("Lista de reproducción borrada")
	return nil
}

// GetSongs devuelve todas las canciones de la lista de reproducción.
func (s *InmemorySongStorage) GetSongs() ([]*voice.Song, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Se copian las canciones para evitar modificaciones inadvertidas.
	songs := make([]*voice.Song, len(s.songs))
	copy(songs, s.songs)

	s.logger.Info("Obteniendo todas las canciones de la lista de reproducción")
	return songs, nil
}

// PopFirstSong elimina y devuelve la primera canción de la lista de reproducción.
func (s *InmemorySongStorage) PopFirstSong() (*voice.Song, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.songs) == 0 {
		s.logger.Info("No hay canciones para eliminar")
		return nil, bot.ErrNoSongs
	}

	song := s.songs[0]
	s.songs = s.songs[1:]
	s.logger.Info("Primera canción eliminada de la lista de reproducción")
	return song, nil
}
