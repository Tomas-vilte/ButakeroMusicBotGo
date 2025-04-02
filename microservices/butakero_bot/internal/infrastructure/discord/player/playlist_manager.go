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

var ErrPlaylistEmpty = errors.New("la playlist está vacía")

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
		pm.logger.Error("No se pudo agregar la canción a la playlist", zap.Error(err))
		return fmt.Errorf("error al agregar la canción: %w", err)
	}
	pm.logger.Debug("Canción agregada exitosamente a la playlist", zap.String("titulo", song.DiscordSong.TitleTrack))
	return nil
}

func (pm *PlaylistManager) RemoveSong(position int) (*entity.DiscordEntity, error) {
	song, err := pm.songStorage.RemoveSong(position)
	if err != nil {
		pm.logger.Error("No se pudo eliminar la canción de la playlist", zap.Error(err), zap.Int("posicion", position))
		return nil, fmt.Errorf("error al eliminar la canción: %w", err)
	}
	pm.logger.Debug("Canción eliminada exitosamente", zap.String("titulo", song.DiscordSong.TitleTrack))
	return song.DiscordSong, nil
}

func (pm *PlaylistManager) GetPlaylist() ([]string, error) {
	songs, err := pm.songStorage.GetSongs()
	if err != nil {
		pm.logger.Error("No se pudo obtener la playlist", zap.Error(err))
		return nil, fmt.Errorf("error al obtener la playlist: %w", err)
	}

	playlist := make([]string, len(songs))
	for i, song := range songs {
		playlist[i] = song.DiscordSong.TitleTrack
	}
	return playlist, nil
}

func (pm *PlaylistManager) ClearPlaylist() error {
	if err := pm.songStorage.ClearPlaylist(); err != nil {
		pm.logger.Error("No se pudo limpiar la playlist", zap.Error(err))
		return fmt.Errorf("error al limpiar la playlist: %w", err)
	}
	pm.logger.Debug("Playlist limpiada exitosamente")
	return nil
}

func (pm *PlaylistManager) GetNextSong() (*entity.PlayedSong, error) {
	song, err := pm.songStorage.PopFirstSong()
	if errors.Is(err, inmemory.ErrNoSongs) {
		return nil, ErrPlaylistEmpty
	}
	if err != nil {
		pm.logger.Error("No se pudo obtener la siguiente canción", zap.Error(err))
		return nil, fmt.Errorf("error al obtener la siguiente canción: %w", err)
	}
	pm.logger.Debug("Reproduciendo siguiente canción", zap.String("titulo", song.DiscordSong.TitleTrack))
	return song, nil
}
