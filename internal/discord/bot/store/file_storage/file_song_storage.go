package file_storage

import (
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"go.uber.org/zap"
	"os"
	"sync"
)

// FileSongStorage implementa la interfaz SongStorage utilizando un archivo para almacenar la lista de reproducción.
type FileSongStorage struct {
	mutex      sync.RWMutex   // mutex se utiliza para garantizar la concurrencia segura al manipular la lista de reproducción.
	filepath   string         // filepath es la ruta al archivo donde se guarda la lista de reproducción.
	logger     logging.Logger // logger es un registrador para registrar mensajes de depuración y errores.
	persistent StatePersistent
}

// NewFileSongStorage crea una nueva instancia de FileSongStorage utilizando el archivo especificado.
// Si el archivo no existe, se creará uno nuevo.
func NewFileSongStorage(filepath string, logger logging.Logger, persistent StatePersistent) (*FileSongStorage, error) {
	// Verificar si el archivo existe, si no existe, crear uno nuevo.
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		if err := os.WriteFile(filepath, []byte("{}"), 0644); err != nil {
			return nil, fmt.Errorf("error al crear el archivo: %w", err)
		}
	}
	return &FileSongStorage{
		filepath:   filepath, // Asignar la ruta del archivo.
		logger:     logger,   // Inicializar un nuevo logger con un logger "Nop" (sin operación) por defecto.
		persistent: persistent,
	}, nil
}

// PrependSong agrega una canción al principio de la lista de reproducción.
func (s *FileSongStorage) PrependSong(song *voice.Song) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.persistent.ReadState(s.filepath)
	if err != nil {
		s.logger.Error("Error al leer el estado", zap.String("filepath", s.filepath), zap.Error(err))
		return err
	}

	state.Songs = append([]*voice.Song{song}, state.Songs...)

	if err := s.persistent.WriteState(s.filepath, state); err != nil {
		s.logger.Error("Error al escribir el estado", zap.String("filepath", s.filepath), zap.Error(err))
		return err
	}

	return nil
}

// AppendSong agrega una canción al final de la lista de reproducción.
func (s *FileSongStorage) AppendSong(song *voice.Song) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.persistent.ReadState(s.filepath)
	if err != nil {
		s.logger.Error("Error al leer el estado", zap.String("filepath", s.filepath), zap.Error(err))
		return err
	}

	state.Songs = append(state.Songs, song)

	if err := s.persistent.WriteState(s.filepath, state); err != nil {
		s.logger.Error("Error al escribir el estado", zap.String("filepath", s.filepath), zap.Error(err))
		return err
	}

	return nil
}

// RemoveSong elimina una canción de la lista de reproducción por posición.
func (s *FileSongStorage) RemoveSong(position int) (*voice.Song, error) {
	index := position - 1

	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.persistent.ReadState(s.filepath)
	if err != nil {
		s.logger.Error("Error al leer el estado", zap.String("filepath", s.filepath), zap.Error(err))
		return nil, err
	}

	if index >= len(state.Songs) || index < 0 {
		s.logger.Error("Posición de canción inválida")
		return nil, bot.ErrRemoveInvalidPosition
	}

	song := state.Songs[index]

	copy(state.Songs[index:], state.Songs[index+1:])
	state.Songs[len(state.Songs)-1] = nil
	state.Songs = state.Songs[:len(state.Songs)-1]

	if err := s.persistent.WriteState(s.filepath, state); err != nil {
		s.logger.Error("Error al escribir el estado", zap.String("filepath", s.filepath), zap.Error(err))
		return nil, err
	}

	return song, nil
}

// ClearPlaylist elimina todas las canciones de la lista de reproducción.
func (s *FileSongStorage) ClearPlaylist() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.persistent.ReadState(s.filepath)
	if err != nil {
		s.logger.Error("Error al leer el estado", zap.String("filepath", s.filepath), zap.Error(err))
		return err
	}

	state.Songs = make([]*voice.Song, 0)

	if err := s.persistent.WriteState(s.filepath, state); err != nil {
		s.logger.Error("Error al escribir el estado", zap.String("filepath", s.filepath), zap.Error(err))
		return err
	}

	return nil
}

// GetSongs devuelve todas las canciones de la lista de reproducción.
func (s *FileSongStorage) GetSongs() ([]*voice.Song, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	state, err := s.persistent.ReadState(s.filepath)
	if err != nil {
		s.logger.Error("Error al leer el estado", zap.String("filepath", s.filepath), zap.Error(err))
		return nil, err
	}

	return state.Songs, nil
}

// PopFirstSong elimina y devuelve la primera canción de la lista de reproducción.
func (s *FileSongStorage) PopFirstSong() (*voice.Song, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.persistent.ReadState(s.filepath)
	if err != nil {
		s.logger.Error("Error al leer el estado", zap.String("filepath", s.filepath), zap.Error(err))
		return nil, err
	}

	if len(state.Songs) == 0 {
		s.logger.Error("No hay canciones disponibles")
		return nil, bot.ErrNoSongs
	}

	song := state.Songs[0]
	state.Songs = state.Songs[1:]

	if err := s.persistent.WriteState(s.filepath, state); err != nil {
		s.logger.Error("Error al escribir las canciones", zap.String("filepath", s.filepath), zap.Error(err))
		return nil, err
	}

	return song, nil
}
