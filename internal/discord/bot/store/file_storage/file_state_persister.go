package file_storage

import (
	"encoding/json"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"go.uber.org/zap"
	"os"
)

type FileState struct {
	Songs        []*voice.Song     `json:"songs"`         // Canciones en la lista de reproducción.
	CurrentSong  *voice.PlayedSong `json:"current_song"`  // Canción actual que se está reproduciendo.
	VoiceChannel string            `json:"voice_channel"` // ID del canal de voz.
	TextChannel  string            `json:"text_channel"`  // ID del canal de texto.
}

type Logger interface {
	Error(msg string, fields ...zap.Field)
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
}

type StatePersistent interface {
	ReadState(filepath string) (*FileState, error)
	WriteState(filepath string, state *FileState) error
}

type JSONStatePersistent struct{}

func NewJSONStatePersistent() *JSONStatePersistent {
	return &JSONStatePersistent{}
}

func (p *JSONStatePersistent) ReadState(filepath string) (*FileState, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var state FileState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	return &state, nil
}

func (p *JSONStatePersistent) WriteState(filepath string, state *FileState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return err
	}
	return nil
}
