package file_storage

import (
	"encoding/json"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestJSONStatePersistent_ReadState(t *testing.T) {
	filepath := "test_state.json"
	testState := &FileState{
		Songs:        []*voice.Song{{Title: "Song 1"}, {Title: "Song 2"}},
		CurrentSong:  &voice.PlayedSong{Song: voice.Song{Title: "Song1"}},
		VoiceChannel: "123456789",
		TextChannel:  "987654321",
	}
	jsonData, err := json.Marshal(testState)
	assert.NoError(t, err)

	err = os.WriteFile(filepath, jsonData, 0644)
	assert.NoError(t, err)

	p := NewJSONStatePersistent()

	readState, err := p.ReadState(filepath)
	assert.NoError(t, err)
	assert.NotNil(t, readState)
	assert.Equal(t, testState, readState)
}

func TestJSONStatePersistent_ReadState_FileReadError(t *testing.T) {
	filepath := "non_existing_file.json"
	p := NewJSONStatePersistent()

	readState, err := p.ReadState(filepath)
	assert.Error(t, err)
	assert.Nil(t, readState)
	assert.True(t, errors.Is(err, os.ErrNotExist))
}

func TestJSONStatePersistent_ReadState_JSONDecodeError(t *testing.T) {
	filepath := "test_state.json"
	invalidData := []byte("invalid json data")
	err := os.WriteFile(filepath, invalidData, 0644)
	assert.NoError(t, err)

	p := NewJSONStatePersistent()

	readState, err := p.ReadState(filepath)
	assert.Error(t, err)
	assert.Nil(t, readState)
}

func TestJSONStatePersistent_WriteState(t *testing.T) {
	filepath := "test_state.json"
	testState := &FileState{
		Songs:        []*voice.Song{{Title: "Song 1"}, {Title: "Song 2"}},
		CurrentSong:  &voice.PlayedSong{Song: voice.Song{Title: "Song1"}},
		VoiceChannel: "123456789",
		TextChannel:  "987654321",
	}

	p := NewJSONStatePersistent()

	err := p.WriteState(filepath, testState)
	assert.NoError(t, err)

	readData, err := os.ReadFile(filepath)
	assert.NoError(t, err)

	var readState FileState
	err = json.Unmarshal(readData, &readState)
	assert.NoError(t, err)
	assert.NotNil(t, readState)
	assert.Equal(t, testState, &readState)
}

func TestJSONStatePersistent_WriteState_Error(t *testing.T) {
	filepath := "non_existing_directory/non_existing_file.json"
	testState := &FileState{
		Songs:        []*voice.Song{{Title: "Song 1"}, {Title: "Song 2"}},
		CurrentSong:  &voice.PlayedSong{Song: voice.Song{Title: "Song1"}},
		VoiceChannel: "123456789",
		TextChannel:  "987654321",
	}

	p := NewJSONStatePersistent()

	err := p.WriteState(filepath, testState)
	assert.Error(t, err)
}
