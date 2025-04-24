package player

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/inmemory"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
)

var ErrPlaylistEmpty = errors.New("la playlist está vacía")

type PlaylistManager struct {
	songStorage ports.PlaylistStorage
	logger      logging.Logger
}

func NewPlaylistManager(songStorage ports.PlaylistStorage, logger logging.Logger) *PlaylistManager {
	return &PlaylistManager{
		songStorage: songStorage,
		logger:      logger,
	}
}

func (pm *PlaylistManager) AddSong(ctx context.Context, song *entity.PlayedSong) error {
	logger := pm.logger.With(
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "AddSong"),
	)

	if song == nil || song.DiscordSong == nil {
		logger.Error("Intento de agregar canción inválida")
		return errors.New("canción inválida")
	}

	if err := pm.songStorage.AppendTrack(ctx, song); err != nil {
		logger.Error("No se pudo agregar la canción a la playlist",
			zap.String("song_id", song.DiscordSong.ID),
			zap.Error(err))
		return fmt.Errorf("error al agregar la canción: %w", err)
	}

	logger.Info("Canción agregada exitosamente",
		zap.String("song_id", song.DiscordSong.ID),
		zap.String("title", song.DiscordSong.TitleTrack))
	return nil
}

func (pm *PlaylistManager) RemoveSong(ctx context.Context, position int) (*entity.DiscordEntity, error) {
	logger := pm.logger.With(
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "RemoveSong"),
		zap.Int("position", position),
	)

	song, err := pm.songStorage.RemoveTrack(ctx, position)
	if err != nil {
		logger.Error("Error al eliminar canción", zap.Error(err))
		return nil, fmt.Errorf("error al eliminar la canción: %w", err)
	}

	logger.Info("Canción eliminada exitosamente",
		zap.String("song_id", song.DiscordSong.ID),
		zap.String("title", song.DiscordSong.TitleTrack))
	return song.DiscordSong, nil
}

func (pm *PlaylistManager) GetPlaylist(ctx context.Context) ([]string, error) {
	logger := pm.logger.With(
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "GetPlaylist"),
	)

	songs, err := pm.songStorage.GetAllTracks(ctx)
	if err != nil {
		logger.Error("Error al obtener playlist", zap.Error(err))
		return nil, fmt.Errorf("error al obtener la playlist: %w", err)
	}

	playlist := make([]string, len(songs))
	for i, song := range songs {
		playlist[i] = song.DiscordSong.TitleTrack
	}

	logger.Debug("Playlist obtenida", zap.Int("count", len(playlist)))
	return playlist, nil
}

func (pm *PlaylistManager) ClearPlaylist(ctx context.Context) error {
	logger := pm.logger.With(
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "ClearPlaylist"),
	)

	if err := pm.songStorage.ClearPlaylist(ctx); err != nil {
		logger.Error("Error al limpiar playlist", zap.Error(err))
		return fmt.Errorf("error al limpiar la playlist: %w", err)
	}

	logger.Info("Playlist limpiada exitosamente")
	return nil
}

func (pm *PlaylistManager) GetNextSong(ctx context.Context) (*entity.PlayedSong, error) {
	logger := pm.logger.With(
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "GetNextSong"),
	)

	song, err := pm.songStorage.PopNextTrack(ctx)
	if err != nil {
		if errors.Is(err, inmemory.ErrNoSongs) {
			logger.Debug("Playlist vacía")
			return nil, ErrPlaylistEmpty
		}
		logger.Error("Error al obtener siguiente canción", zap.Error(err))
		return nil, fmt.Errorf("error al obtener la siguiente canción: %w", err)
	}

	logger.Info("Siguiente canción obtenida",
		zap.String("song_id", song.DiscordSong.ID),
		zap.String("title", song.DiscordSong.TitleTrack))
	return song, nil
}
