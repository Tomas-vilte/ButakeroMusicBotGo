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
		return fmt.Errorf("error al agregar la canción: %w", err)
	}

	gp.eventCh <- PlayerEvent{
		Type: EventPlay,
		Payload: EventPayload{
			TextChannelID:  textChannelID,
			VoiceChannelID: voiceChannelID,
		},
	}

	gp.logger.Debug("Canción agregada a la lista", zap.String("título", playedSong.DiscordSong.TitleTrack))
	return nil
}

func (gp *GuildPlayer) SkipSong() {
	gp.playbackHandler.Stop()
	gp.logger.Debug("Se salteó la canción actual")
}

func (gp *GuildPlayer) Pause() error {
	if err := gp.playbackHandler.Pause(); err != nil {
		gp.logger.Error("Error al pausar la reproducción", zap.Error(err))
		return err
	}
	gp.logger.Debug("Reproducción pausada")
	return nil
}

func (gp *GuildPlayer) Resume() error {
	if err := gp.playbackHandler.Resume(); err != nil {
		gp.logger.Error("Error al reanudar la reproducción", zap.Error(err))
		return err
	}
	gp.logger.Debug("Reproducción reanudada")
	return nil
}

func (gp *GuildPlayer) Stop() error {
	if err := gp.playlistHandler.ClearPlaylist(); err != nil {
		gp.logger.Error("Error al limpiar la lista de reproducción", zap.Error(err))
		return fmt.Errorf("error al limpiar la lista: %w", err)
	}

	gp.playbackHandler.Stop()
	gp.logger.Debug("Reproducción detenida y lista limpiada")
	return nil
}

func (gp *GuildPlayer) RemoveSong(position int) (*entity.DiscordEntity, error) {
	song, err := gp.playlistHandler.RemoveSong(position)
	if err != nil {
		return nil, fmt.Errorf("error al remover la cancion: %w", err)
	}

	gp.logger.Debug("Se removio la cancion de la playlist", zap.String("cancion", song.TitleTrack))
	return song, nil
}

func (gp *GuildPlayer) GetPlaylist() ([]string, error) {
	playlist, err := gp.playlistHandler.GetPlaylist()
	if err != nil {
		return nil, fmt.Errorf("error al obtener la playlist: %w", err)
	}

	gp.logger.Debug("Lista de reproducción obtenida", zap.Int("cantidad", len(playlist)))
	return playlist, nil
}

func (gp *GuildPlayer) GetPlayedSong() (*entity.PlayedSong, error) {
	currentSong, err := gp.stateStorage.GetCurrentSong()
	if err != nil {
		gp.logger.Error("Error al obtener la canción actual", zap.Error(err))
		return nil, fmt.Errorf("error al obtener la cancion actual: %w", err)
	}
	return currentSong, nil
}

func (gp *GuildPlayer) Run(ctx context.Context) error {
	gp.mu.Lock()
	if gp.running {
		gp.mu.Unlock()
		return errors.New("el reproductor ya está ejecutándose")
	}
	gp.running = true
	gp.mu.Unlock()

	defer func() {
		gp.mu.Lock()
		gp.running = false
		gp.mu.Unlock()
	}()

	currentSong, err := gp.stateStorage.GetCurrentSong()
	if err != nil {
		gp.logger.Error("Error al obtener el estado de la canción actual", zap.Error(err))
	}

	if currentSong != nil {
		currentSong.StartPosition += currentSong.Position
		if err := gp.playlistHandler.AddSong(currentSong); err != nil {
			gp.logger.Error("Error al restaurar la canción actual a la lista", zap.Error(err))
		}
	}

	for {
		select {
		case <-ctx.Done():
			gp.logger.Info("Cerrando el reproductor")
			return nil

		case event := <-gp.eventCh:
			if err := gp.handleEvent(ctx, event); err != nil {
				gp.logger.Error("Error al manejar el evento del reproductor",
					zap.String("evento", event.Type),
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
		return fmt.Errorf("tipo de evento desconocido: %s", event.Type)
	}
}

func (gp *GuildPlayer) handlePlayEvent(ctx context.Context, event PlayerEvent) error {
	payload, ok := event.Payload.(EventPayload)
	if !ok {
		return errors.New("payload de evento para reproduccion no valida")
	}

	if payload.TextChannelID != nil {
		if err := gp.stateStorage.SetTextChannel(*payload.TextChannelID); err != nil {
			return fmt.Errorf("error al setear el canal de texto: %w", err)
		}
	}
	if payload.VoiceChannelID != nil {
		if err := gp.stateStorage.SetVoiceChannel(*payload.VoiceChannelID); err != nil {
			return fmt.Errorf("error al setear el canal de voz: %w", err)
		}
	}

	voiceChannel, textChannel, err := gp.getVoiceAndTextChannels()
	if err != nil {
		return fmt.Errorf("error al obtener los canales: %w", err)
	}

	if err := gp.JoinVoiceChannel(voiceChannel); err != nil {
		return fmt.Errorf("error al unirse al canal de voz: %w", err)
	}

	for {
		song, err := gp.playlistHandler.GetNextSong()
		if errors.Is(err, ErrPlaylistEmpty) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error al obtener la siguente cancion: %w", err)
		}

		if err := gp.playbackHandler.Play(ctx, song, textChannel); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			gp.logger.Error("Error al reproducir la canción", zap.Error(err))
			continue
		}

		for gp.playbackHandler.CurrentState() != StateIdle {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (gp *GuildPlayer) getVoiceAndTextChannels() (voiceChannel string, textChannel string, err error) {
	voiceChannel, err = gp.stateStorage.GetVoiceChannel()
	if err != nil {
		return "", "", fmt.Errorf("error al obtener el canal de voz: %w", err)
	}
	textChannel, err = gp.stateStorage.GetTextChannel()
	if err != nil {
		return "", "", fmt.Errorf("error al obtener el canal de texto: %w", err)
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
