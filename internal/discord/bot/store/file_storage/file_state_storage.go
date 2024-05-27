package file_storage

import (
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"go.uber.org/zap"
	"os"
	"sync"
)

// FileStateStorage implementa la interfaz StateStorage utilizando un archivo para almacenar el estado del reproductor.
type FileStateStorage struct {
	mutex      sync.RWMutex   // mutex se utiliza para garantizar la concurrencia segura al manipular el estado del reproductor.
	filepath   string         // filepath es la ruta al archivo donde se guarda el estado.
	logger     logging.Logger // logger es un registrador para registrar mensajes de depuración y errores.
	persistent StatePersistent
}

// NewFileStateStorage crea una nueva instancia de FileStateStorage utilizando el archivo especificado.
// Si el archivo no existe, se creará uno nuevo.
func NewFileStateStorage(filepath string, logger logging.Logger, persistent StatePersistent) (*FileStateStorage, error) {
	// Verificar si el archivo existe, si no existe, crear uno nuevo.
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		if err := os.WriteFile(filepath, []byte("{}"), 0644); err != nil {
			return nil, fmt.Errorf("error al crear el archivo: %w", err)
		}
	}
	return &FileStateStorage{
		filepath:   filepath, // Asignar la ruta del archivo.
		logger:     logger,   // Inicializar un nuevo logger
		persistent: persistent,
	}, nil
}

// GetCurrentSong devuelve la canción actual que se está reproduciendo.
func (s *FileStateStorage) GetCurrentSong() (*voice.PlayedSong, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	state, err := s.persistent.ReadState(s.filepath)
	if err != nil {
		s.logger.Error("Error al leer el estado", zap.Error(err))
		return nil, err
	}

	return state.CurrentSong, nil
}

// SetCurrentSong establece la canción actual que se está reproduciendo.
func (s *FileStateStorage) SetCurrentSong(song *voice.PlayedSong) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.persistent.ReadState(s.filepath)
	if err != nil {
		s.logger.Error("Error al leer el estado", zap.Error(err))
		return err
	}

	state.CurrentSong = song

	if err := s.persistent.WriteState(s.filepath, state); err != nil {
		s.logger.Error("Error al escribir el estado", zap.Error(err))
		return err
	}

	return nil
}

// GetVoiceChannel devuelve el ID del canal de voz.
func (s *FileStateStorage) GetVoiceChannel() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	state, err := s.persistent.ReadState(s.filepath)
	if err != nil {
		s.logger.Error("Error al leer el estado", zap.Error(err))
		return "", err
	}

	return state.VoiceChannel, nil
}

// SetVoiceChannel establece el ID del canal de voz.
func (s *FileStateStorage) SetVoiceChannel(channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.persistent.ReadState(s.filepath)
	if err != nil {
		s.logger.Error("Error al leer el estado", zap.Error(err))
		return err
	}

	state.VoiceChannel = channelID

	if err := s.persistent.WriteState(s.filepath, state); err != nil {
		s.logger.Error("Error al escribir el estado", zap.Error(err))
		return err
	}

	return nil
}

// GetTextChannel devuelve el ID del canal de texto.
func (s *FileStateStorage) GetTextChannel() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	state, err := s.persistent.ReadState(s.filepath)
	if err != nil {
		s.logger.Error("Error al leer el estado", zap.Error(err))
		return "", err
	}

	return state.TextChannel, nil
}

// SetTextChannel establece el ID del canal de texto.
func (s *FileStateStorage) SetTextChannel(channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.persistent.ReadState(s.filepath)
	if err != nil {
		s.logger.Error("Error al leer el estado", zap.Error(err))
		return err
	}

	state.TextChannel = channelID

	if err := s.persistent.WriteState(s.filepath, state); err != nil {
		s.logger.Error("Error al escribir el estado", zap.Error(err))
		return err
	}
	return nil
}
