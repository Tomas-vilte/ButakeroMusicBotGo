//go:build !integration

package player

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/inmemory"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestAddSong(t *testing.T) {
	// arrange
	mockStorage := new(MockPlaylistStorage)
	logger := new(logging.MockLogger)
	pm := NewPlaylistManager(mockStorage, logger)

	ctx := context.Background()
	testSong := &entity.PlayedSong{
		DiscordSong: &entity.DiscordEntity{
			ID:         "123",
			TitleTrack: "Test Song",
		},
	}

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	mockStorage.On("AppendTrack", ctx, testSong).Return(nil)

	// act
	err := pm.AddSong(ctx, testSong)

	// assert
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

func TestAddSong_InvalidSong(t *testing.T) {
	// arrange
	mockStorage := new(MockPlaylistStorage)
	logger := new(logging.MockLogger)
	pm := NewPlaylistManager(mockStorage, logger)

	ctx := context.Background()

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()

	// act
	err := pm.AddSong(ctx, nil)

	// assert
	assert.Error(t, err)
	assert.Equal(t, "canción inválida", err.Error())
	mockStorage.AssertNotCalled(t, "AppendTrack")
}

func TestRemoveSong(t *testing.T) {
	// arrange
	mockStorage := new(MockPlaylistStorage)
	logger := new(logging.MockLogger)
	pm := NewPlaylistManager(mockStorage, logger)

	ctx := context.Background()
	position := 0
	expectedSong := &entity.PlayedSong{
		DiscordSong: &entity.DiscordEntity{
			ID:         "123",
			TitleTrack: "Test Song",
		},
	}

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	mockStorage.On("RemoveTrack", ctx, position).Return(expectedSong, nil)

	// act
	song, err := pm.RemoveSong(ctx, position)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, expectedSong.DiscordSong, song)
	mockStorage.AssertExpectations(t)
}

func TestRemoveSong_Error(t *testing.T) {
	// arrange
	mockStorage := new(MockPlaylistStorage)
	logger := new(logging.MockLogger)
	pm := NewPlaylistManager(mockStorage, logger)

	ctx := context.Background()
	position := 0
	expectedErr := errors.New("error de almacenamiento")

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()
	mockStorage.On("RemoveTrack", ctx, position).Return((*entity.PlayedSong)(nil), expectedErr)

	// act
	song, err := pm.RemoveSong(ctx, position)

	// assert
	assert.Error(t, err)
	assert.Nil(t, song)
	assert.Contains(t, err.Error(), "error al eliminar la canción")
	mockStorage.AssertExpectations(t)
}

func TestGetPlaylist(t *testing.T) {
	// arrange
	mockStorage := new(MockPlaylistStorage)
	logger := new(logging.MockLogger)
	pm := NewPlaylistManager(mockStorage, logger)

	ctx := context.Background()
	expectedSongs := []*entity.PlayedSong{
		{
			DiscordSong: &entity.DiscordEntity{
				TitleTrack: "Song 1",
			},
		},
		{
			DiscordSong: &entity.DiscordEntity{
				TitleTrack: "Song 2",
			},
		},
	}

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()
	mockStorage.On("GetAllTracks", ctx).Return(expectedSongs, nil)

	// act
	playlist, err := pm.GetPlaylist(ctx)

	// assert
	assert.NoError(t, err)
	assert.Len(t, playlist, 2)
	assert.Equal(t, "Song 1", playlist[0])
	assert.Equal(t, "Song 2", playlist[1])
	mockStorage.AssertExpectations(t)
}

func TestGetPlaylist_Error(t *testing.T) {
	// arrange
	mockStorage := new(MockPlaylistStorage)
	logger := new(logging.MockLogger)
	pm := NewPlaylistManager(mockStorage, logger)

	ctx := context.Background()
	expectedErr := errors.New("error de almacenamiento")

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()
	mockStorage.On("GetAllTracks", ctx).Return(([]*entity.PlayedSong)(nil), expectedErr)

	// act
	playlist, err := pm.GetPlaylist(ctx)

	// assert
	assert.Error(t, err)
	assert.Nil(t, playlist)
	assert.Contains(t, err.Error(), "error al obtener la playlist")
	mockStorage.AssertExpectations(t)
}

func TestClearPlaylist(t *testing.T) {
	// arrange
	mockStorage := new(MockPlaylistStorage)
	logger := new(logging.MockLogger)
	pm := NewPlaylistManager(mockStorage, logger)

	ctx := context.Background()

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	mockStorage.On("ClearPlaylist", ctx).Return(nil)

	// act
	err := pm.ClearPlaylist(ctx)

	// assert
	assert.NoError(t, err)
	mockStorage.AssertExpectations(t)
}

func TestClearPlaylist_Error(t *testing.T) {
	// arrange
	mockStorage := new(MockPlaylistStorage)
	logger := new(logging.MockLogger)
	pm := NewPlaylistManager(mockStorage, logger)

	ctx := context.Background()
	expectedErr := errors.New("error de almacenamiento")

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()
	mockStorage.On("ClearPlaylist", ctx).Return(expectedErr)

	// act
	err := pm.ClearPlaylist(ctx)

	// assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al limpiar la playlist")
	mockStorage.AssertExpectations(t)
}

func TestGetNextSong(t *testing.T) {
	// arrange
	mockStorage := new(MockPlaylistStorage)
	logger := new(logging.MockLogger)
	pm := NewPlaylistManager(mockStorage, logger)

	ctx := context.Background()
	expectedSong := &entity.PlayedSong{
		DiscordSong: &entity.DiscordEntity{
			ID:         "123",
			TitleTrack: "Next Song",
		},
	}

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	mockStorage.On("PopNextTrack", ctx).Return(expectedSong, nil)

	// act
	song, err := pm.GetNextSong(ctx)

	// assert
	assert.NoError(t, err)
	assert.Equal(t, expectedSong, song)
	mockStorage.AssertExpectations(t)
}

func TestGetNextSong_EmptyPlaylist(t *testing.T) {
	// arrange
	mockStorage := new(MockPlaylistStorage)
	logger := new(logging.MockLogger)
	pm := NewPlaylistManager(mockStorage, logger)

	ctx := context.Background()

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	mockStorage.On("PopNextTrack", ctx).Return((*entity.PlayedSong)(nil), inmemory.ErrNoSongs)

	// act
	song, err := pm.GetNextSong(ctx)

	// assert
	assert.Error(t, err)
	assert.Nil(t, song)
	assert.Equal(t, ErrPlaylistEmpty, err)
	mockStorage.AssertExpectations(t)
}

func TestGetNextSong_Error(t *testing.T) {
	// arrange
	mockStorage := new(MockPlaylistStorage)
	logger := new(logging.MockLogger)
	pm := NewPlaylistManager(mockStorage, logger)

	ctx := context.Background()
	expectedErr := errors.New("error de almacenamiento")

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Error", mock.Anything, mock.Anything).Return()
	mockStorage.On("PopNextTrack", ctx).Return((*entity.PlayedSong)(nil), expectedErr)

	// act
	song, err := pm.GetNextSong(ctx)

	// assert
	assert.Error(t, err)
	assert.Nil(t, song)
	assert.Contains(t, err.Error(), "error al obtener la siguiente canción")
	mockStorage.AssertExpectations(t)
}
