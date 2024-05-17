package file_storage

import (
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestFileSongStorage_PrependSong(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	song := &voice.Song{Title: "Test Song"}

	initialState := &FileState{
		Songs: []*voice.Song{},
	}

	mockPersistent.On("ReadState", filepath).Return(initialState, nil)
	mockPersistent.On("WriteState", filepath, mock.AnythingOfType("*file_storage.FileState")).Return(nil)

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	err = storage.PrependSong(song)
	assert.NoError(t, err)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_PrependSong_ReadStateError(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	song := &voice.Song{Title: "Test Song"}
	readError := errors.New("error al leer el estado")

	// Configuramos el mock para devolver un *FileState vacío y un error.
	mockPersistent.On("ReadState", filepath).Return(&FileState{}, readError)
	mockLogger.On("Error", "Error al leer el estado", mock.Anything).Return()

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	err = storage.PrependSong(song)
	assert.Error(t, err)
	assert.Equal(t, readError.Error(), err.Error())

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_PrependSong_WriteStateError(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	song := &voice.Song{Title: "Test Song"}
	initialState := &FileState{
		Songs: []*voice.Song{},
	}
	writeError := errors.New("error al escribir el estado")

	mockPersistent.On("ReadState", filepath).Return(initialState, nil)
	mockPersistent.On("WriteState", filepath, mock.AnythingOfType("*file_storage.FileState")).Return(writeError)
	mockLogger.On("Error", "Error al escribir el estado", mock.Anything).Return()

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	err = storage.PrependSong(song)
	assert.Error(t, err)
	assert.Equal(t, writeError, err)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_AppendSong(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	song := &voice.Song{Title: "Test Song"}

	initialState := &FileState{
		Songs: []*voice.Song{},
	}

	mockPersistent.On("ReadState", filepath).Return(initialState, nil)
	mockPersistent.On("WriteState", filepath, mock.AnythingOfType("*file_storage.FileState")).Return(nil)

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	err = storage.AppendSong(song)
	assert.NoError(t, err)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_AppendSong_ReadStateError(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	song := &voice.Song{Title: "Test Song"}
	readError := errors.New("error al leer el estado")
	initialState := &FileState{
		Songs: []*voice.Song{},
	}

	mockPersistent.On("ReadState", filepath).Return(initialState, readError)
	mockLogger.On("Error", "Error al leer el estado", mock.Anything).Return()

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	err = storage.AppendSong(song)
	assert.Error(t, err)
	assert.Equal(t, readError, err)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_AppendSong_WriteStateError(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	song := &voice.Song{Title: "Test Song"}
	initialState := &FileState{
		Songs: []*voice.Song{},
	}
	writeError := errors.New("error al escribir el estado")

	mockPersistent.On("ReadState", filepath).Return(initialState, nil)
	mockPersistent.On("WriteState", filepath, mock.AnythingOfType("*file_storage.FileState")).Return(writeError)
	mockLogger.On("Error", "Error al escribir el estado", mock.Anything).Return()

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	err = storage.AppendSong(song)
	assert.Error(t, err)
	assert.Equal(t, writeError, err)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_RemoveSong(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	song := &voice.Song{Title: "Test Song"}

	initialState := &FileState{
		Songs: []*voice.Song{song},
	}

	mockPersistent.On("ReadState", filepath).Return(initialState, nil)
	mockPersistent.On("WriteState", filepath, mock.AnythingOfType("*file_storage.FileState")).Return(nil)

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	removedSong, err := storage.RemoveSong(1)
	assert.NoError(t, err)
	assert.Equal(t, song, removedSong)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_RemoveSong_ReadStateError(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	readError := errors.New("error al leer el estado")
	initialState := &FileState{}

	mockPersistent.On("ReadState", filepath).Return(initialState, readError)
	mockLogger.On("Error", "Error al leer el estado", mock.Anything).Return()

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	_, err = storage.RemoveSong(1)
	assert.Error(t, err)
	assert.Equal(t, readError, err)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_RemoveSong_InvalidPosition(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	initialState := &FileState{
		Songs: []*voice.Song{},
	}

	mockPersistent.On("ReadState", filepath).Return(initialState, nil)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	_, err = storage.RemoveSong(1)
	assert.Error(t, err)
	assert.Equal(t, bot.ErrRemoveInvalidPosition, err)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_RemoveSong_WriteStateError(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	song := &voice.Song{Title: "Test Song"}
	initialState := &FileState{
		Songs: []*voice.Song{song},
	}
	writeError := errors.New("error al escribir el estado")

	mockPersistent.On("ReadState", filepath).Return(initialState, nil)
	mockPersistent.On("WriteState", filepath, mock.AnythingOfType("*file_storage.FileState")).Return(writeError)
	mockLogger.On("Error", "Error al escribir el estado", mock.Anything).Return()

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	_, err = storage.RemoveSong(1)
	assert.Error(t, err)
	assert.Equal(t, writeError, err)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_ClearPlaylist(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"

	initialState := &FileState{
		Songs: []*voice.Song{{Title: "Test Song"}},
	}

	mockPersistent.On("ReadState", filepath).Return(initialState, nil)
	mockPersistent.On("WriteState", filepath, mock.AnythingOfType("*file_storage.FileState")).Return(nil)

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	err = storage.ClearPlaylist()
	assert.NoError(t, err)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_ClearPlaylist_ReadStateError(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	readError := errors.New("error al leer el estado")
	initialState := &FileState{
		Songs: []*voice.Song{{Title: "Test Song"}},
	}

	mockPersistent.On("ReadState", filepath).Return(initialState, readError)
	mockLogger.On("Error", "Error al leer el estado", mock.Anything).Return()

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	err = storage.ClearPlaylist()
	assert.Error(t, err)
	assert.Equal(t, readError, err)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_ClearPlaylist_WriteStateError(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	initialState := &FileState{
		Songs: []*voice.Song{{Title: "Test Song"}},
	}
	writeError := errors.New("error al escribir el estado")

	mockPersistent.On("ReadState", filepath).Return(initialState, nil)
	mockPersistent.On("WriteState", filepath, mock.AnythingOfType("*file_storage.FileState")).Return(writeError)
	mockLogger.On("Error", "Error al escribir el estado", mock.Anything).Return()

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	err = storage.ClearPlaylist()
	assert.Error(t, err)
	assert.Equal(t, writeError, err)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_GetSongs(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	song := &voice.Song{Title: "Test Song"}

	initialState := &FileState{
		Songs: []*voice.Song{song},
	}

	mockPersistent.On("ReadState", filepath).Return(initialState, nil)

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	songs, err := storage.GetSongs()
	assert.NoError(t, err)
	assert.Equal(t, []*voice.Song{song}, songs)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_GetSongs_ReadStateError(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	readError := errors.New("error al leer el estado")
	song := &voice.Song{Title: "Test Song"}

	initialState := &FileState{
		Songs: []*voice.Song{song},
	}

	mockPersistent.On("ReadState", filepath).Return(initialState, readError)
	mockLogger.On("Error", "Error al leer el estado", mock.Anything).Return()

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	_, err = storage.GetSongs()
	assert.Error(t, err)
	assert.Equal(t, readError, err)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_PopFirstSong(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	song := &voice.Song{Title: "Test Song"}

	initialState := &FileState{
		Songs: []*voice.Song{song},
	}

	mockPersistent.On("ReadState", filepath).Return(initialState, nil)
	mockPersistent.On("WriteState", filepath, mock.AnythingOfType("*file_storage.FileState")).Return(nil)

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	poppedSong, err := storage.PopFirstSong()
	assert.NoError(t, err)
	assert.Equal(t, song, poppedSong)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_PopFirstSong_ReadStateError(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	readError := errors.New("error al leer el estado")
	song := &voice.Song{Title: "Test Song"}

	initialState := &FileState{
		Songs: []*voice.Song{song},
	}

	mockPersistent.On("ReadState", filepath).Return(initialState, readError)
	mockLogger.On("Error", "Error al leer el estado", mock.Anything).Return()

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	_, err = storage.PopFirstSong()
	assert.Error(t, err)
	assert.Equal(t, readError, err)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_PopFirstSong_WriteStateError(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	song := &voice.Song{Title: "Test Song"}
	initialState := &FileState{
		Songs: []*voice.Song{song},
	}
	writeError := errors.New("error al escribir el estado")

	mockPersistent.On("ReadState", filepath).Return(initialState, nil)
	mockPersistent.On("WriteState", filepath, mock.AnythingOfType("*file_storage.FileState")).Return(writeError)
	mockLogger.On("Error", "Error al escribir las canciones", mock.Anything).Return()

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	_, err = storage.PopFirstSong()
	assert.Error(t, err)
	assert.Equal(t, writeError, err)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestFileSongStorage_PopFirstSong_NoSongs(t *testing.T) {
	mockLogger := new(MockLogger)
	mockPersistent := new(MockStatePersistent)

	filepath := "test_state.json"
	initialState := &FileState{
		Songs: []*voice.Song{},
	}

	mockPersistent.On("ReadState", filepath).Return(initialState, nil)
	mockLogger.On("Error", "No hay canciones disponibles", mock.Anything).Return()

	storage, err := NewFileSongStorage(filepath, mockLogger, mockPersistent)
	assert.NoError(t, err)

	_, err = storage.PopFirstSong()
	assert.Error(t, err)
	assert.Equal(t, bot.ErrNoSongs, err)

	mockPersistent.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
