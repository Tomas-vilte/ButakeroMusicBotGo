package player

import (
	"errors"
	"fmt"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/inmemory"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
)

type PlaylistManager struct {
	songStorage ports.SongStorage
	logger      logging.Logger
}

func NewPlaylistManager(songStorage ports.SongStorage, logger logging.Logger) *PlaylistManager {
	return &PlaylistManager{
		songStorage: songStorage,
		logger:      logger,
	}
}

func (pm *PlaylistManager) AddSong(song *entity.PlayedSong) error {
	if err := pm.songStorage.AppendSong(song); err != nil {
		pm.logger.Error("Error adding song to playlist", zap.Error(err))
		return fmt.Errorf("error adding song: %w", err)
	}
	return nil
}

func (pm *PlaylistManager) RemoveSong(position int) (*entity.DiscordEntity, error) {
	song, err := pm.songStorage.RemoveSong(position)
	if err != nil {
		pm.logger.Error("Error removing song from playlist", zap.Error(err), zap.Int("position", position))
		return nil, fmt.Errorf("error removing song: %w", err)
	}
	return song.DiscordSong, nil
}

func (pm *PlaylistManager) GetPlaylist() ([]string, error) {
	songs, err := pm.songStorage.GetSongs()
	if err != nil {
		pm.logger.Error("Error getting playlist", zap.Error(err))
		return nil, fmt.Errorf("error getting playlist: %w", err)
	}

	playlist := make([]string, len(songs))
	for i, song := range songs {
		playlist[i] = song.DiscordSong.TitleTrack
	}
	return playlist, nil
}

func (pm *PlaylistManager) ClearPlaylist() error {
	if err := pm.songStorage.ClearPlaylist(); err != nil {
		pm.logger.Error("Error clearing playlist", zap.Error(err))
		return fmt.Errorf("error clearing playlist: %w", err)
	}
	return nil
}

func (pm *PlaylistManager) GetNextSong() (*entity.PlayedSong, error) {
	song, err := pm.songStorage.PopFirstSong()
	if errors.Is(err, inmemory.ErrNoSongs) {
		return nil, ErrPlaylistEmpty
	}
	if err != nil {
		pm.logger.Error("Error getting next song", zap.Error(err))
		return nil, fmt.Errorf("error getting next song: %w", err)
	}
	return song, nil
}

var ErrPlaylistEmpty = errors.New("playlist is empty")
