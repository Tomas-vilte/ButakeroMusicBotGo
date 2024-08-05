package file_storage

import (
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestFileStateStorage_GetCurrentSong(t *testing.T) {
	filepath := "test_state.json"
	testState := &FileState{
		CurrentSong: &voice.PlayedSong{Song: voice.Song{Title: "Song1"}},
	}
	persistent := NewJSONStatePersistent()
	err := persistent.WriteState(filepath, testState)
	if err != nil {
		return
	}

	storage, err := NewFileStateStorage(filepath, nil, persistent)
	assert.NoError(t, err)

	song, err := storage.GetCurrentSong()
	assert.NoError(t, err)
	assert.NotNil(t, song)
	assert.Equal(t, testState.CurrentSong, song)
}

func TestFileStateStorage_GetCurrentSong_Error(t *testing.T) {
	// Simulate error from ReadState
	filepath := "test_state.json"
	mockLogger := new(logging.MockLogger)
	mockPersistent := new(MockStatePersistent)

	// Simulate error returned by ReadState
	mockPersistent.On("ReadState", filepath).Return(&FileState{}, errors.New("error reading state"))
	mockLogger.On("Error", "Error al leer el estado", mock.Anything).Return()

	storage, err := NewFileStateStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	song, err := storage.GetCurrentSong()
	assert.Error(t, err)
	assert.Nil(t, song)
}

func TestFileStateStorage_SetCurrentSong(t *testing.T) {
	filepath := "test_state.json"
	initialState := &FileState{
		Songs: []*voice.Song{
			{Title: "Dua lipa"},
			{Title: "Ke personaje"},
		},
	}
	persistent := NewJSONStatePersistent()
	err := persistent.WriteState(filepath, initialState)
	if err != nil {
		return
	}

	storage, err := NewFileStateStorage(filepath, nil, persistent)
	assert.NoError(t, err)

	newSong := &voice.PlayedSong{Song: voice.Song{Title: "New Song"}}
	err = storage.SetCurrentSong(newSong)
	assert.NoError(t, err)

	song, err := storage.GetCurrentSong()
	assert.NoError(t, err)
	assert.NotNil(t, song)
	assert.Equal(t, newSong, song)
}

func TestFileStateStorage_SetCurrentSong_Error(t *testing.T) {
	filepath := "test_state.json"
	mockLogger := new(logging.MockLogger)
	mockPersistent := new(MockStatePersistent)

	mockPersistent.On("ReadState", filepath).Return(&FileState{}, errors.New("error reading state"))
	mockLogger.On("Error", "Error al leer el estado", mock.Anything).Return()

	storage, err := NewFileStateStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	newSong := &voice.PlayedSong{Song: voice.Song{Title: "New Song"}}
	err = storage.SetCurrentSong(newSong)
	assert.Error(t, err)
}

func TestFileStateStorage_GetVoiceChannel(t *testing.T) {
	filepath := "test_state.json"
	testState := &FileState{
		VoiceChannel: "123456789",
	}
	persistent := NewJSONStatePersistent()
	err := persistent.WriteState(filepath, testState)
	if err != nil {
		return
	}

	storage, err := NewFileStateStorage(filepath, nil, persistent)
	assert.NoError(t, err)

	channelID, err := storage.GetVoiceChannel()
	assert.NoError(t, err)
	assert.Equal(t, testState.VoiceChannel, channelID)
}

func TestFileStateStorage_GetVoiceChannel_Error(t *testing.T) {
	filepath := "test_state.json"
	mockLogger := new(logging.MockLogger)
	mockPersistent := new(MockStatePersistent)

	// Simulate error returned by ReadState
	mockPersistent.On("ReadState", filepath).Return(&FileState{}, errors.New("error reading state"))
	mockLogger.On("Error", "Error al leer el estado", mock.Anything).Return()

	storage, err := NewFileStateStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	channelID, err := storage.GetVoiceChannel()
	assert.Error(t, err)
	assert.Equal(t, "", channelID)
}

func TestFileStateStorage_SetVoiceChannel(t *testing.T) {
	filepath := "test_state.json"
	initialState := &FileState{}
	persistent := NewJSONStatePersistent()
	err := persistent.WriteState(filepath, initialState)
	if err != nil {
		return
	}

	storage, err := NewFileStateStorage(filepath, nil, persistent)
	assert.NoError(t, err)

	newChannelID := "987654321"
	err = storage.SetVoiceChannel(newChannelID)
	assert.NoError(t, err)

	channelID, err := storage.GetVoiceChannel()
	assert.NoError(t, err)
	assert.Equal(t, newChannelID, channelID)
}

func TestFileStateStorage_SetTextChannel_Error(t *testing.T) {
	filepath := "test_state.json"
	mockLogger := new(logging.MockLogger)
	mockPersistent := new(MockStatePersistent)

	mockPersistent.On("ReadState", filepath).Return(&FileState{}, errors.New("error reading state"))
	mockLogger.On("Error", "Error al leer el estado", mock.Anything).Return()

	storage, err := NewFileStateStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	newChannelID := "123456789"
	err = storage.SetTextChannel(newChannelID)
	assert.Error(t, err)
}

func TestFileStateStorage_GetTextChannel(t *testing.T) {
	filepath := "test_state.json"
	testState := &FileState{
		TextChannel: "987654321",
	}
	persistent := NewJSONStatePersistent()
	err := persistent.WriteState(filepath, testState)
	if err != nil {
		return
	}

	storage, err := NewFileStateStorage(filepath, nil, persistent)
	assert.NoError(t, err)

	channelID, err := storage.GetTextChannel()
	assert.NoError(t, err)
	assert.Equal(t, testState.TextChannel, channelID)
}

func TestFileStateStorage_SetTextChannel(t *testing.T) {
	filepath := "test_state.json"
	initialState := &FileState{}
	persistent := NewJSONStatePersistent()
	err := persistent.WriteState(filepath, initialState)
	if err != nil {
		return
	}

	storage, err := NewFileStateStorage(filepath, nil, persistent)
	assert.NoError(t, err)

	newChannelID := "123456789"
	err = storage.SetTextChannel(newChannelID)
	assert.NoError(t, err)

	channelID, err := storage.GetTextChannel()
	assert.NoError(t, err)
	assert.Equal(t, newChannelID, channelID)
}

func TestFileStateStorage_SetVoiceChannel_Write_Error(t *testing.T) {
	filepath := "test_state.json"
	mockLogger := new(logging.MockLogger)
	mockPersistent := new(MockStatePersistent)
	initialState := &FileState{
		Songs: []*voice.Song{
			{Title: "Dua lipa"},
		},
	}

	mockPersistent.On("ReadState", filepath).Return(initialState, nil)
	mockPersistent.On("WriteState", filepath, mock.Anything).Return(errors.New("error writing state"))
	mockLogger.On("Error", "Error al escribir el estado", mock.Anything).Return()

	storage, err := NewFileStateStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	newChannelID := "987654321"
	err = storage.SetVoiceChannel(newChannelID)
	assert.Error(t, err)
}

func TestFileStateStorage_SetCurrentSong_Write_Error(t *testing.T) {
	filepath := "test_state.json"
	mockLogger := new(logging.MockLogger)
	mockPersistent := new(MockStatePersistent)
	initialState := &FileState{
		Songs: []*voice.Song{
			{Title: "Dua lipa"},
		},
	}
	writeError := errors.New("error al escribir el estado")

	mockPersistent.On("ReadState", filepath).Return(initialState, nil)

	// Simulate error returned by WriteState
	mockPersistent.On("WriteState", filepath, mock.AnythingOfType("*file_storage.FileState")).Return(writeError)
	mockLogger.On("Error", "Error al escribir el estado", mock.Anything).Return()

	storage, err := NewFileStateStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	newSong := &voice.PlayedSong{Song: voice.Song{Title: "New Song"}}
	err = storage.SetCurrentSong(newSong)
	assert.Error(t, err)
}

func TestFileStateStorage_GetTextChannel_Read_Error(t *testing.T) {
	filepath := "test_state.json"
	mockLogger := new(logging.MockLogger)
	mockPersistent := new(MockStatePersistent)
	readError := errors.New("error al leer el estado")
	song := &voice.Song{Title: "Test Song"}

	initialState := &FileState{
		Songs: []*voice.Song{song},
	}

	// Simulate error returned by ReadState
	mockPersistent.On("ReadState", filepath).Return(initialState, readError)
	mockLogger.On("Error", "Error al leer el estado", mock.Anything).Return()

	storage, err := NewFileStateStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	channelID, err := storage.GetTextChannel()
	assert.Error(t, err)
	assert.Equal(t, "", channelID)
}

func TestFileStateStorage_SetVoiceChannel_Read_Error(t *testing.T) {
	filepath := "test_state.json"
	mockLogger := new(logging.MockLogger)
	mockPersistent := new(MockStatePersistent)
	readError := errors.New("error al leer el estado")
	song := &voice.Song{Title: "Test Song"}

	initialState := &FileState{
		Songs: []*voice.Song{song},
	}

	// Simulate error returned by WriteState
	mockPersistent.On("ReadState", filepath).Return(initialState, readError)
	mockLogger.On("Error", "Error al leer el estado", mock.Anything).Return()

	storage, err := NewFileStateStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	newChannelID := "123456789"
	err = storage.SetVoiceChannel(newChannelID)
	assert.Error(t, err)
}

func TestFileStateStorage_SetTextChannel_Read_Error(t *testing.T) {
	filepath := "test_state.json"
	mockLogger := new(logging.MockLogger)
	mockPersistent := new(MockStatePersistent)
	readError := errors.New("error al leer el estado")
	song := &voice.Song{Title: "Test Song"}

	initialState := &FileState{
		Songs: []*voice.Song{song},
	}

	// Simulate error returned by WriteState
	mockPersistent.On("ReadState", filepath).Return(initialState, readError)
	mockLogger.On("Error", "Error al leer el estado", mock.Anything).Return()

	storage, err := NewFileStateStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	newChannelID := "123456789"
	err = storage.SetTextChannel(newChannelID)
	assert.Error(t, err)
}

func TestFileStateStorage_SetTextChannel_Write_Error(t *testing.T) {
	filepath := "test_state.json"
	mockLogger := new(logging.MockLogger)
	mockPersistent := new(MockStatePersistent)
	initialState := &FileState{
		Songs: []*voice.Song{
			{Title: "Dua lipa"},
		},
	}
	writeError := errors.New("error al escribir el estado")
	mockPersistent.On("ReadState", filepath).Return(initialState, nil)

	mockPersistent.On("WriteState", filepath, mock.Anything).Return(writeError)
	mockLogger.On("Error", "Error al escribir el estado", mock.Anything).Return()

	storage, err := NewFileStateStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	newChannelID := "123456789"
	err = storage.SetTextChannel(newChannelID)
	assert.Error(t, err)
}
