//go:build !integration

package inmemory

import (
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reflect"
	"testing"
)

func TestInmemorySongStorage_PrependSong(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	storage := NewInmemorySongStorage(mockLogger)
	song := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "Test Song"}}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	err := storage.PrependSong(song)

	assert.NoError(t, err)
	mockLogger.AssertCalled(t, "Info", "Canción agregada al principio de la lista de reproducción", mock.AnythingOfType("[]zapcore.Field"))
}

func TestInmemorySongStorage_AppendSong(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	storage := NewInmemorySongStorage(mockLogger)
	song := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "Test Song"}}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	err := storage.AppendSong(song)

	assert.NoError(t, err)
	mockLogger.AssertCalled(t, "Info", "Canción agregada al final de la lista de reproducción", mock.AnythingOfType("[]zapcore.Field"))
}

func TestInmemorySongStorage_RemoveSong(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	storage := NewInmemorySongStorage(mockLogger)

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	song1 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "1"}}
	song2 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "2"}}
	song3 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "3"}}

	err := storage.AppendSong(song1)
	if err != nil {
		t.Errorf("Error al agregar canción: %v", err)
	}
	err = storage.AppendSong(song2)
	if err != nil {
		t.Errorf("Error al agregar canción: %v", err)
	}
	err = storage.AppendSong(song3)
	if err != nil {
		t.Errorf("Error al agregar canción: %v", err)
	}

	// Test case: Eliminar canción en posición válida
	removedSong, err := storage.RemoveSong(2)
	if err != nil {
		t.Errorf("Error al eliminar canción: %v", err)
	}
	if removedSong.DiscordSong.TitleTrack != song2.DiscordSong.TitleTrack {
		t.Errorf("Canción incorrecta eliminada. Esperado: %s, Obtenido: %s", song2.DiscordSong.TitleTrack, removedSong.DiscordSong.TitleTrack)
	}

	// Test case: Eliminar canción en posición inválida
	_, err = storage.RemoveSong(0)
	if !errors.Is(err, ErrRemoveInvalidPosition) {
		t.Errorf("Error esperado: %v, Error obtenido: %v", ErrRemoveInvalidPosition, err)
	}

	// Test case: Eliminar canción en posición inválida (fuera de rango)
	_, err = storage.RemoveSong(4)
	if !errors.Is(err, ErrRemoveInvalidPosition) {
		t.Errorf("Error esperado: %v, Error obtenido: %v", ErrRemoveInvalidPosition, err)
	}

	mockLogger.AssertExpectations(t)
}

func TestInmemorySongStorage_GetSongs(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	storage := NewInmemorySongStorage(mockLogger)

	song1 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "1"}}
	song2 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "2"}}
	song3 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "3"}}

	err := storage.AppendSong(song1)
	if err != nil {
		t.Errorf("Error al agregar canción: %v", err)
	}
	err = storage.AppendSong(song2)
	if err != nil {
		t.Errorf("Error al agregar canción: %v", err)
	}
	err = storage.AppendSong(song3)
	if err != nil {
		t.Errorf("Error al agregar canción: %v", err)
	}

	// Test case: Obtener todas las canciones de la lista de reproducción
	songs, err := storage.GetSongs()
	if err != nil {
		t.Errorf("Error al obtener canciones: %v", err)
	}

	expectedSongs := []*entity.PlayedSong{song1, song2, song3}
	if !reflect.DeepEqual(songs, expectedSongs) {
		t.Errorf("Canciones obtenidas incorrectas. Esperado: %v, Obtenido: %v", expectedSongs, songs)
	}

	mockLogger.AssertExpectations(t)
}

func TestInmemorySongStorage_PopFirstSong(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	storage := NewInmemorySongStorage(mockLogger)

	song1 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "1"}}
	err := storage.AppendSong(song1)
	if err != nil {
		t.Errorf("Error al agregar canción: %v", err)
	}

	// Test case: Eliminar la primera canción de la lista de reproducción
	poppedSong, err := storage.PopFirstSong()
	if err != nil {
		t.Errorf("Error al eliminar la primera canción: %v", err)
	}
	if poppedSong.DiscordSong.TitleTrack != song1.DiscordSong.TitleTrack {
		t.Errorf("Canción incorrecta eliminada. Esperado: %s, Obtenido: %s", song1.DiscordSong.TitleTrack, poppedSong.DiscordSong.TitleTrack)
	}

	// Test case: Intentar eliminar canción de una lista de reproducción vacía
	_, err = storage.PopFirstSong()
	if !errors.Is(err, ErrNoSongs) {
		t.Errorf("Error esperado: %v, Error obtenido: %v", ErrNoSongs, err)
	}

	mockLogger.AssertExpectations(t)
}

func TestInmemorySongStorage_ClearPlaylist(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	storage := NewInmemorySongStorage(mockLogger)

	// Agregar algunas canciones a la lista de reproducción
	song1 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "1"}}
	err := storage.AppendSong(song1)
	if err != nil {
		return
	}

	// Llamar al método ClearPlaylist para eliminar todas las canciones
	err = storage.ClearPlaylist()
	if err != nil {
		t.Errorf("Error al borrar la lista de reproducción: %v", err)
	}

	// Verificar que la lista de reproducción esté vacía
	songs, _ := storage.GetSongs()
	if len(songs) != 0 {
		t.Errorf("La lista de reproducción no está vacía después de borrarla")
	}

	mockLogger.AssertExpectations(t)
}
