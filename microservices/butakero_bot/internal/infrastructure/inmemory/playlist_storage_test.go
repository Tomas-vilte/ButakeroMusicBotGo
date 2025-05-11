//go:build !integration

package inmemory

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/errors_app"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"reflect"
	"sync"
	"testing"
	"time"
)

var errRemoveInvalidPosition = errors_app.NewAppError(errors_app.ErrCodeInvalidTrackPosition, "Posición de la canción inválida", nil)
var errPlaylistEmpty = errors_app.NewAppError(errors_app.ErrCodePlaylistEmpty, "No hay canciones disponibles en la playlist", nil)

func TestInmemorySongStorage_AppendSong(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	storage := NewInmemoryPlaylistStorage(mockLogger)
	song := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "Test Song"}}

	ctx := context.Background()

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	err := storage.AppendTrack(ctx, song)

	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)
}

func TestInmemorySongStorage_RemoveSong(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	storage := NewInmemoryPlaylistStorage(mockLogger)

	ctx := context.Background()

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	song1 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "1"}}
	song2 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "2"}}
	song3 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "3"}}

	err := storage.AppendTrack(ctx, song1)
	if err != nil {
		t.Errorf("Error al agregar canción: %v", err)
	}
	err = storage.AppendTrack(ctx, song2)
	if err != nil {
		t.Errorf("Error al agregar canción: %v", err)
	}
	err = storage.AppendTrack(ctx, song3)
	if err != nil {
		t.Errorf("Error al agregar canción: %v", err)
	}

	// Test case: Eliminar canción en posición válida
	removedSong, err := storage.RemoveTrack(ctx, 2)
	if err != nil {
		t.Errorf("Error al eliminar canción: %v", err)
	}
	if removedSong.DiscordSong.TitleTrack != song2.DiscordSong.TitleTrack {
		t.Errorf("Canción incorrecta eliminada. Esperado: %s, Obtenido: %s", song2.DiscordSong.TitleTrack, removedSong.DiscordSong.TitleTrack)
	}

	// Test case: Eliminar canción en posición inválida
	_, err = storage.RemoveTrack(ctx, -1)
	if !compareErrors(t, errRemoveInvalidPosition, err) {
		t.Errorf("Error esperado: %v, Error obtenido: %v", errRemoveInvalidPosition, err)
	}

	// Test case: Eliminar canción en posición inválida (fuera de rango)
	_, err = storage.RemoveTrack(ctx, 100)
	if !compareErrors(t, errRemoveInvalidPosition, err) {
		t.Errorf("Error esperado: %v, Error obtenido: %v", errRemoveInvalidPosition, err)
	}

	mockLogger.AssertExpectations(t)
}

func compareErrors(t *testing.T, expected, got error) bool {
	var expectedAppErr *errors_app.AppError
	var gotAppErr *errors_app.AppError

	if !errors.As(expected, &expectedAppErr) || !errors.As(got, &gotAppErr) {
		t.Errorf("Uno o ambos errores no son del tipo AppError")
		return false
	}

	if expectedAppErr.Code != gotAppErr.Code {
		t.Errorf("Códigos de error diferentes. Esperado: %v, Obtenido: %v",
			expectedAppErr.Code, gotAppErr.Code)
		return false
	}

	return true
}

func TestPlaylistConcurrency(t *testing.T) {
	mockLogger := new(logging.MockLogger)

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	storage := NewInmemoryPlaylistStorage(mockLogger)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			song := &entity.PlayedSong{
				DiscordSong: &entity.DiscordEntity{ID: fmt.Sprintf("song%d", i), AddedAt: time.Now()},
			}
			_ = storage.AppendTrack(context.Background(), song)
		}(i)
	}

	wg.Wait()

	if len(storage.songs) != 100 {
		t.Errorf("Expected 100 songs, got %d", len(storage.songs))
	}

	for i := 1; i < len(storage.songs); i++ {
		if storage.songs[i-1].DiscordSong.AddedAt.After(storage.songs[i].DiscordSong.AddedAt) {
			t.Errorf("Playlist not in chronological order")
		}
	}
}

func TestInmemorySongStorage_GetSongs(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	storage := NewInmemoryPlaylistStorage(mockLogger)

	ctx := context.Background()

	song1 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "1"}}
	song2 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "2"}}
	song3 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "3"}}

	err := storage.AppendTrack(ctx, song1)
	if err != nil {
		t.Errorf("Error al agregar canción: %v", err)
	}
	err = storage.AppendTrack(ctx, song2)
	if err != nil {
		t.Errorf("Error al agregar canción: %v", err)
	}
	err = storage.AppendTrack(ctx, song3)
	if err != nil {
		t.Errorf("Error al agregar canción: %v", err)
	}

	// Test case: Obtener todas las canciones de la lista de reproducción
	songs, err := storage.GetAllTracks(ctx)
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
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	storage := NewInmemoryPlaylistStorage(mockLogger)

	ctx := context.Background()

	song1 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "1"}}
	err := storage.AppendTrack(ctx, song1)
	if err != nil {
		t.Errorf("Error al agregar canción: %v", err)
	}

	// Test case: Eliminar la primera canción de la lista de reproducción
	poppedSong, err := storage.PopNextTrack(ctx)
	if err != nil {
		t.Errorf("Error al eliminar la primera canción: %v", err)
	}
	if poppedSong.DiscordSong.TitleTrack != song1.DiscordSong.TitleTrack {
		t.Errorf("Canción incorrecta eliminada. Esperado: %s, Obtenido: %s", song1.DiscordSong.TitleTrack, poppedSong.DiscordSong.TitleTrack)
	}

	// Test case: Intentar eliminar canción de una lista de reproducción vacía
	_, err = storage.PopNextTrack(ctx)
	if !compareErrors(t, errPlaylistEmpty, err) {
		t.Errorf("Error esperado: %v, Error obtenido: %v", errPlaylistEmpty, err)
	}

	mockLogger.AssertExpectations(t)
}

func TestInmemorySongStorage_ClearPlaylist(t *testing.T) {
	mockLogger := new(logging.MockLogger)
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	storage := NewInmemoryPlaylistStorage(mockLogger)

	ctx := context.Background()

	// Agregar algunas canciones a la lista de reproducción
	song1 := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "1"}}
	err := storage.AppendTrack(ctx, song1)
	if err != nil {
		return
	}

	// Llamar al método ClearPlaylist para eliminar todas las canciones
	err = storage.ClearPlaylist(ctx)
	if err != nil {
		t.Errorf("Error al borrar la lista de reproducción: %v", err)
	}

	// Verificar que la lista de reproducción esté vacía
	songs, _ := storage.GetAllTracks(ctx)
	if len(songs) != 0 {
		t.Errorf("La lista de reproducción no está vacía después de borrarla")
	}

	mockLogger.AssertExpectations(t)
}
