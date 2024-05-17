package inmemory_storage

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot/store/file_storage"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"sync"
)

// InmemoryStateStorage implementa la interfaz StateStorage utilizando la memoria RAM para almacenar el estado del reproductor.
type InmemoryStateStorage struct {
	mutex        sync.RWMutex        // mutex se utiliza para garantizar la concurrencia segura al manipular el estado del reproductor.
	songs        []*voice.Song       // songs es la lista de reproducción de canciones almacenada en memoria.
	currentSong  *voice.PlayedSong   // currentSong es la canción actual que se está reproduciendo.
	textChannel  string              // textChannel es el ID del canal de texto asociado al servidor.
	voiceChannel string              // voiceChannel es el ID del canal de voz asociado al servidor.
	logger       file_storage.Logger // logger es un registrador para registrar mensajes de depuración y errores.
}

// NewInmemoryStateStorage crea una nueva instancia de InmemoryStateStorage.
func NewInmemoryStateStorage(logger file_storage.Logger) *InmemoryStateStorage {
	return &InmemoryStateStorage{
		mutex:  sync.RWMutex{},         // Se inicializa un nuevo mutex para garantizar la concurrencia segura.
		songs:  make([]*voice.Song, 0), // Se inicializa una nueva lista de reproducción de canciones vacía.
		logger: logger,                 // Se inicializa un nuevo logger con un logger "Nop" (sin operación) por defecto.
	}
}

// GetCurrentSong devuelve la canción actual que se está reproduciendo.
func (s *InmemoryStateStorage) GetCurrentSong() (*voice.PlayedSong, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.logger.Debug("Obteniendo la canción actual")
	return s.currentSong, nil
}

// SetCurrentSong establece la canción actual que se está reproduciendo.
func (s *InmemoryStateStorage) SetCurrentSong(song *voice.PlayedSong) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.currentSong = song
	s.logger.Debug("Canción actual establecida")
	return nil
}

// GetVoiceChannel devuelve el ID del canal de voz.
func (s *InmemoryStateStorage) GetVoiceChannel() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.logger.Debug("Obteniendo el canal de voz")
	return s.voiceChannel, nil
}

// SetVoiceChannel establece el ID del canal de voz.
func (s *InmemoryStateStorage) SetVoiceChannel(channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.voiceChannel = channelID
	s.logger.Debug("Canal de voz establecido")
	return nil
}

// GetTextChannel devuelve el ID del canal de texto.
func (s *InmemoryStateStorage) GetTextChannel() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	s.logger.Debug("Obteniendo el canal de texto")
	return s.textChannel, nil
}

// SetTextChannel establece el ID del canal de texto.
func (s *InmemoryStateStorage) SetTextChannel(channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.textChannel = channelID
	s.logger.Debug("Canal de texto establecido")
	return nil
}
