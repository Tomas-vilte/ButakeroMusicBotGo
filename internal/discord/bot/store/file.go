package store

import (
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"os"
	"sync"
)

// fileState que representa el estado de la lista de reproduccion
type fileState struct {
	Songs        []*bot.Song     `json:"songs"`         // Lista de canciones
	CurrentSong  *bot.PlayedSong `json:"current_song"`  // Canción actual que se está reproduciendo
	VoiceChannel string          `json:"voice_channel"` // ID del canal de voz donde está conectado el bot
	TextChannel  string          `json:"text_channel"`  // ID del canal de texto donde los usuarios interactúan con el bot
}

// FilePlaylistStorage para el almacenamiento de listas de reproducción en archivos
type FilePlaylistStorage struct {
	mutex    sync.RWMutex // Mutex para acceso seguro a los datos
	filepath string       // Ruta del archivo donde se almacena la lista de reproducción
}

// NewFilePlaylistStorage Crea una nueva instancia de FilePlaylistStorage
func NewFilePlaylistStorage(filePath string) (*FilePlaylistStorage, error) {
	// Verifica si el archivo de la lista de reproducción existe
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Si no existe, crea un archivo vacío
		if err := os.WriteFile(filePath, []byte("{}"), 0644); err != nil {
			return nil, fmt.Errorf("error en crear el archivo: %w", err)
		}
	}
	// Crea la instancia de FilePlaylistStorage
	return &FilePlaylistStorage{
		mutex:    sync.RWMutex{},
		filepath: filePath,
	}, nil
}

// GetCurrentSong Obtiene la canción actual que se está reproduciendo
func (s *FilePlaylistStorage) GetCurrentSong() (*bot.PlayedSong, error) {
	// Bloquea el mutex para lectura
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Lee el estado de la lista de reproducción
	state, err := s.readState()
	if err != nil {
		return nil, fmt.Errorf("failed to read state: %w", err)
	}

	// Retorna la canción actual
	return state.CurrentSong, nil
}

// SetCurrentSong Establece la canción actual que se está reproduciendo
func (s *FilePlaylistStorage) SetCurrentSong(song *bot.PlayedSong) error {
	// Bloquea el mutex para escritura
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Lee el estado de la lista de reproducción
	state, err := s.readState()
	if err != nil {
		return fmt.Errorf("failed to read state: %w", err)
	}

	// Actualiza la canción actual
	state.CurrentSong = song

	// Escribe el estado actualizado de la lista de reproducción
	if err := s.writeState(state); err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	return nil
}

// GetVoiceChannel Obtiene el ID del canal de voz donde está conectado el bot
func (s *FilePlaylistStorage) GetVoiceChannel() (string, error) {
	// Bloquea el mutex para lectura
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Lee el estado de la lista de reproducción
	state, err := s.readState()
	if err != nil {
		return "", fmt.Errorf("failed to read state: %w", err)
	}

	// Retorna el ID del canal de voz
	return state.VoiceChannel, nil
}

func (s *FilePlaylistStorage) SetVoiceChannel(channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.readState()
	if err != nil {
		return fmt.Errorf("failed to read state: %w", err)
	}

	state.VoiceChannel = channelID

	if err := s.writeState(state); err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	return nil
}

func (s *FilePlaylistStorage) GetTextChannel() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	state, err := s.readState()
	if err != nil {
		return "", fmt.Errorf("failed to read state: %w", err)
	}

	return state.TextChannel, nil
}

func (s *FilePlaylistStorage) SetTextChannel(channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.readState()
	if err != nil {
		return fmt.Errorf("failed to read state: %w", err)
	}

	state.TextChannel = channelID

	if err := s.writeState(state); err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	return nil
}

func (s *FilePlaylistStorage) PrependSong(song *bot.Song) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.readState()
	if err != nil {
		return fmt.Errorf("failed to read state: %w", err)
	}

	state.Songs = append([]*bot.Song{song}, state.Songs...)

	if err := s.writeState(state); err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	return nil
}

func (s *FilePlaylistStorage) AppendSong(song *bot.Song) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.readState()
	if err != nil {
		return fmt.Errorf("failed to read state: %w", err)
	}

	state.Songs = append(state.Songs, song)

	if err := s.writeState(state); err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	return nil
}

func (s *FilePlaylistStorage) RemoveSong(position int) (*bot.Song, error) {
	index := position - 1

	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.readState()
	if err != nil {
		return nil, fmt.Errorf("failed to read state: %w", err)
	}

	if index >= len(state.Songs) || index < 0 {
		return nil, bot.ErrRemoveInvalidPosition
	}

	song := state.Songs[index]

	copy(state.Songs[index:], state.Songs[index+1:])
	state.Songs[len(state.Songs)-1] = nil
	state.Songs = state.Songs[:len(state.Songs)-1]

	if err := s.writeState(state); err != nil {
		return nil, fmt.Errorf("failed to write state: %w", err)
	}

	return song, nil
}

func (s *FilePlaylistStorage) ClearPlaylist() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.readState()
	if err != nil {
		return fmt.Errorf("failed to read state: %w", err)
	}

	state.Songs = make([]*bot.Song, 0)

	if err := s.writeState(state); err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	return nil
}

func (s *FilePlaylistStorage) GetSongs() ([]*bot.Song, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	state, err := s.readState()
	if err != nil {
		return nil, fmt.Errorf("failed to read state: %w", err)
	}

	return state.Songs, nil
}

func (s *FilePlaylistStorage) PopFirstSong() (*bot.Song, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.readState()
	if err != nil {
		return nil, fmt.Errorf("failed to read songs: %w", err)
	}

	if len(state.Songs) == 0 {
		return nil, bot.ErrNoSongs
	}

	song := state.Songs[0]
	state.Songs = state.Songs[1:]

	if err := s.writeState(state); err != nil {
		return nil, fmt.Errorf("failed to write songs: %w", err)
	}

	return song, nil
}

func (s *FilePlaylistStorage) readState() (*fileState, error) {
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var state fileState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal songs: %w", err)
	}

	return &state, nil
}

func (s *FilePlaylistStorage) writeState(state *fileState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(s.filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
