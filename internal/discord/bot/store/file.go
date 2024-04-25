package store

import (
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"os"
	"sync"
)

// fileState representa el estado almacenado en el archivo.
type fileState struct {
	Songs        []*bot.Song     `json:"songs"`         // Canciones en la lista de reproducción.
	CurrentSong  *bot.PlayedSong `json:"current_song"`  // Canción actual que se está reproduciendo.
	VoiceChannel string          `json:"voice_channel"` // ID del canal de voz.
	TextChannel  string          `json:"text_channel"`  // ID del canal de texto.
}

// FilePlaylistStorage es una implementación de GuildPlayerState que almacena la lista de reproducción en un archivo.
type FilePlaylistStorage struct {
	mutex    sync.RWMutex
	filepath string
}

// NewFilePlaylistStorage crea una nueva instancia de FilePlaylistStorage con la ruta de archivo proporcionada.
func NewFilePlaylistStorage(filepath string) (*FilePlaylistStorage, error) {
	// Si el archivo no existe, se crea con un objeto JSON vacío.
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		if err := os.WriteFile(filepath, []byte("{}"), 0644); err != nil {
			return nil, fmt.Errorf("error al crear el archivo: %w", err)
		}
	}

	return &FilePlaylistStorage{
		mutex:    sync.RWMutex{},
		filepath: filepath,
	}, nil
}

// GetCurrentSong devuelve la canción actual que se está reproduciendo.
func (s *FilePlaylistStorage) GetCurrentSong() (*bot.PlayedSong, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	state, err := s.readState()
	if err != nil {
		return nil, fmt.Errorf("error al leer el estado: %w", err)
	}

	return state.CurrentSong, nil
}

// SetCurrentSong establece la canción actual que se está reproduciendo.
func (s *FilePlaylistStorage) SetCurrentSong(song *bot.PlayedSong) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.readState()
	if err != nil {
		return fmt.Errorf("error al leer el estado: %w", err)
	}

	state.CurrentSong = song

	if err := s.writeState(state); err != nil {
		return fmt.Errorf("error al escribir el estado: %w", err)
	}

	return nil
}

// GetVoiceChannel devuelve el ID del canal de voz.
func (s *FilePlaylistStorage) GetVoiceChannel() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	state, err := s.readState()
	if err != nil {
		return "", fmt.Errorf("error al leer el estado: %w", err)
	}

	return state.VoiceChannel, nil
}

// SetVoiceChannel establece el ID del canal de voz.
func (s *FilePlaylistStorage) SetVoiceChannel(channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.readState()
	if err != nil {
		return fmt.Errorf("error al leer el estado: %w", err)
	}

	state.VoiceChannel = channelID

	if err := s.writeState(state); err != nil {
		return fmt.Errorf("error al escribir el estado: %w", err)
	}

	return nil
}

// GetTextChannel devuelve el ID del canal de texto.
func (s *FilePlaylistStorage) GetTextChannel() (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	state, err := s.readState()
	if err != nil {
		return "", fmt.Errorf("error al leer el estado: %w", err)
	}

	return state.TextChannel, nil
}

// SetTextChannel establece el ID del canal de texto.
func (s *FilePlaylistStorage) SetTextChannel(channelID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.readState()
	if err != nil {
		return fmt.Errorf("error al leer el estado: %w", err)
	}

	state.TextChannel = channelID

	if err := s.writeState(state); err != nil {
		return fmt.Errorf("error al escribir el estado: %w", err)
	}

	return nil
}

// PrependSong agrega una canción al principio de la lista de reproducción.
func (s *FilePlaylistStorage) PrependSong(song *bot.Song) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.readState()
	if err != nil {
		return fmt.Errorf("error al leer el estado: %w", err)
	}

	state.Songs = append([]*bot.Song{song}, state.Songs...)

	if err := s.writeState(state); err != nil {
		return fmt.Errorf("error al escribir el estado: %w", err)
	}

	return nil
}

// AppendSong agrega una canción al final de la lista de reproducción.
func (s *FilePlaylistStorage) AppendSong(song *bot.Song) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.readState()
	if err != nil {
		return fmt.Errorf("error al leer el estado: %w", err)
	}

	state.Songs = append(state.Songs, song)

	if err := s.writeState(state); err != nil {
		return fmt.Errorf("error al escribir el estado: %w", err)
	}

	return nil
}

// RemoveSong elimina una canción de la lista de reproducción por posición.
func (s *FilePlaylistStorage) RemoveSong(position int) (*bot.Song, error) {
	index := position - 1

	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.readState()
	if err != nil {
		return nil, fmt.Errorf("error al leer el estado: %w", err)
	}

	if index >= len(state.Songs) || index < 0 {
		return nil, bot.ErrRemoveInvalidPosition
	}

	song := state.Songs[index]

	copy(state.Songs[index:], state.Songs[index+1:])
	state.Songs[len(state.Songs)-1] = nil
	state.Songs = state.Songs[:len(state.Songs)-1]

	if err := s.writeState(state); err != nil {
		return nil, fmt.Errorf("error al escribir el estado: %w", err)
	}

	return song, nil
}

// ClearPlaylist elimina todas las canciones de la lista de reproducción.
func (s *FilePlaylistStorage) ClearPlaylist() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.readState()
	if err != nil {
		return fmt.Errorf("error al leer el estado: %w", err)
	}

	state.Songs = make([]*bot.Song, 0)

	if err := s.writeState(state); err != nil {
		return fmt.Errorf("error al escribir el estado: %w", err)
	}

	return nil
}

// GetSongs devuelve todas las canciones de la lista de reproducción.
func (s *FilePlaylistStorage) GetSongs() ([]*bot.Song, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	state, err := s.readState()
	if err != nil {
		return nil, fmt.Errorf("error al leer el estado: %w", err)
	}

	return state.Songs, nil
}

// PopFirstSong elimina y devuelve la primera canción de la lista de reproducción.
func (s *FilePlaylistStorage) PopFirstSong() (*bot.Song, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	state, err := s.readState()
	if err != nil {
		return nil, fmt.Errorf("error al leer las canciones: %w", err)
	}

	if len(state.Songs) == 0 {
		return nil, bot.ErrNoSongs
	}

	song := state.Songs[0]
	state.Songs = state.Songs[1:]

	if err := s.writeState(state); err != nil {
		return nil, fmt.Errorf("error al escribir las canciones: %w", err)
	}

	return song, nil
}

// readState lee el estado del archivo.
func (s *FilePlaylistStorage) readState() (*fileState, error) {
	data, err := os.ReadFile(s.filepath)
	if err != nil {
		return nil, fmt.Errorf("error al leer el archivo: %w", err)
	}

	var state fileState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("error al deserializar las canciones: %w", err)
	}

	return &state, nil
}

// writeState escribe el estado en el archivo.
func (s *FilePlaylistStorage) writeState(state *fileState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("error al serializar el estado: %w", err)
	}

	if err := os.WriteFile(s.filepath, data, 0644); err != nil {
		return fmt.Errorf("error al escribir el archivo: %w", err)
	}
	return nil
}
