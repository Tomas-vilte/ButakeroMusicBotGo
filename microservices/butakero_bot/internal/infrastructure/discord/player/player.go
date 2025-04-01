package player

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
)

type GuildPlayer struct {
	playbackHandler PlaybackHandler
	playlistHandler PlaylistHandler
	stateStorage    ports.StateStorage
	eventCh         chan PlayerEvent
	logger          logging.Logger
	mu              sync.Mutex
	running         bool
}

func NewGuildPlayer(cfg Config) *GuildPlayer {
	playbackCtrl := NewPlaybackController(
		cfg.VoiceSession,
		cfg.StorageAudio,
		cfg.StateStorage,
		cfg.Messenger,
		cfg.Logger,
	)

	playlistMgr := NewPlaylistManager(cfg.SongStorage, cfg.Logger)

	return &GuildPlayer{
		playbackHandler: playbackCtrl,
		playlistHandler: playlistMgr,
		stateStorage:    cfg.StateStorage,
		eventCh:         make(chan PlayerEvent, 100),
		logger:          cfg.Logger,
	}
}

func (gp *GuildPlayer) AddSong(textChannelID, voiceChannelID *string, playedSong *entity.PlayedSong) error {
	if err := gp.playlistHandler.AddSong(playedSong); err != nil {
		return fmt.Errorf("failed to add song: %w", err)
	}

	gp.eventCh <- PlayerEvent{
		Type: EventPlay,
		Payload: EventPayload{
			TextChannelID:  textChannelID,
			VoiceChannelID: voiceChannelID,
		},
	}

	gp.logger.Info("Song added to playlist", zap.String("title", playedSong.DiscordSong.TitleTrack))
	return nil
}

func (gp *GuildPlayer) SkipSong() {
	gp.playbackHandler.Stop()
	gp.logger.Info("Current song skipped")
}

func (gp *GuildPlayer) Pause() error {
	if err := gp.playbackHandler.Pause(); err != nil {
		gp.logger.Error("Error pausing playback", zap.Error(err))
		return err
	}
	gp.logger.Info("Playback paused")
	return nil
}

func (gp *GuildPlayer) Resume() error {
	if err := gp.playbackHandler.Resume(); err != nil {
		gp.logger.Error("Error resuming playback", zap.Error(err))
		return err
	}
	gp.logger.Info("Playback resumed")
	return nil
}

func (gp *GuildPlayer) Stop() error {
	if err := gp.playlistHandler.ClearPlaylist(); err != nil {
		gp.logger.Error("Error clearing playlist", zap.Error(err))
		return fmt.Errorf("error clearing playlist: %w", err)
	}

	gp.playbackHandler.Stop()
	gp.logger.Info("Playback stopped and playlist cleared")
	return nil
}

func (gp *GuildPlayer) RemoveSong(position int) (*entity.DiscordEntity, error) {
	song, err := gp.playlistHandler.RemoveSong(position)
	if err != nil {
		return nil, fmt.Errorf("error removing song: %w", err)
	}

	gp.logger.Info("Song removed from playlist", zap.String("title", song.TitleTrack))
	return song, nil
}

func (gp *GuildPlayer) GetPlaylist() ([]string, error) {
	playlist, err := gp.playlistHandler.GetPlaylist()
	if err != nil {
		return nil, fmt.Errorf("error getting playlist: %w", err)
	}

	gp.logger.Info("Playlist retrieved", zap.Int("count", len(playlist)))
	return playlist, nil
}

func (gp *GuildPlayer) GetPlayedSong() (*entity.PlayedSong, error) {
	currentSong, err := gp.stateStorage.GetCurrentSong()
	if err != nil {
		gp.logger.Error("Error getting current song", zap.Error(err))
		return nil, fmt.Errorf("error getting current song: %w", err)
	}
	return currentSong, nil
}

func (gp *GuildPlayer) Run(ctx context.Context) error {
	gp.mu.Lock()
	if gp.running {
		gp.mu.Unlock()
		return errors.New("player is already running")
	}
	gp.running = true
	gp.mu.Unlock()

	defer func() {
		gp.mu.Lock()
		gp.running = false
		gp.mu.Unlock()
	}()

	// Restaurar estado inicial
	currentSong, err := gp.stateStorage.GetCurrentSong()
	if err != nil {
		gp.logger.Error("Error getting current song state", zap.Error(err))
	}

	if currentSong != nil {
		currentSong.StartPosition += currentSong.Position
		if err := gp.playlistHandler.AddSong(currentSong); err != nil {
			gp.logger.Error("Error restoring current song to playlist", zap.Error(err))
		}
	}

	for {
		select {
		case <-ctx.Done():
			gp.logger.Info("Player context cancelled, shutting down")
			return nil

		case event := <-gp.eventCh:
			if err := gp.handleEvent(ctx, event); err != nil {
				gp.logger.Error("Error handling player event",
					zap.String("event", event.Type),
					zap.Error(err))
			}
		}
	}
}

func (gp *GuildPlayer) handleEvent(ctx context.Context, event PlayerEvent) error {
	switch event.Type {
	case EventPlay:
		return gp.handlePlayEvent(ctx, event)
	case EventPause:
		return gp.Pause()
	case EventResume:
		return gp.Resume()
	case EventStop:
		return gp.Stop()
	case EventSkip:
		gp.SkipSong()
		return nil
	default:
		return fmt.Errorf("unknown event type: %s", event.Type)
	}
}

func (gp *GuildPlayer) handlePlayEvent(ctx context.Context, event PlayerEvent) error {
	payload, ok := event.Payload.(EventPayload)
	if !ok {
		return errors.New("invalid play event payload")
	}

	// Actualizar canales si se proporcionaron
	if payload.TextChannelID != nil {
		if err := gp.stateStorage.SetTextChannel(*payload.TextChannelID); err != nil {
			return fmt.Errorf("error setting text channel: %w", err)
		}
	}
	if payload.VoiceChannelID != nil {
		if err := gp.stateStorage.SetVoiceChannel(*payload.VoiceChannelID); err != nil {
			return fmt.Errorf("error setting voice channel: %w", err)
		}
	}

	// Obtener canales actuales
	voiceChannel, textChannel, err := gp.getVoiceAndTextChannels()
	if err != nil {
		return fmt.Errorf("error getting channels: %w", err)
	}

	if err := gp.JoinVoiceChannel(voiceChannel); err != nil {
		return fmt.Errorf("error joining voice channel: %w", err)
	}

	// Reproducir la lista de reproducción
	for {
		song, err := gp.playlistHandler.GetNextSong()
		if errors.Is(err, ErrPlaylistEmpty) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error getting next song: %w", err)
		}

		if err := gp.playbackHandler.Play(ctx, song, textChannel); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			gp.logger.Error("Error playing song", zap.Error(err))
			continue
		}

		// Esperar a que termine la canción o sea interrumpida
		for gp.playbackHandler.CurrentState() != StateIdle {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (gp *GuildPlayer) getVoiceAndTextChannels() (voiceChannel string, textChannel string, err error) {
	voiceChannel, err = gp.stateStorage.GetVoiceChannel()
	if err != nil {
		return "", "", fmt.Errorf("error getting voice channel: %w", err)
	}
	textChannel, err = gp.stateStorage.GetTextChannel()
	if err != nil {
		return "", "", fmt.Errorf("error getting text channel: %w", err)
	}
	return voiceChannel, textChannel, nil
}

func (gp *GuildPlayer) JoinVoiceChannel(channelID string) error {
	voiceSession := gp.playbackHandler.GetVoiceSession()
	return voiceSession.JoinVoiceChannel(channelID)
}

func (gp *GuildPlayer) StateStorage() ports.StateStorage {
	return gp.stateStorage
}
