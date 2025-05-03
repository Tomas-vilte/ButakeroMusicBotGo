package player

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/interfaces"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/voice"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
)

type Config struct {
	VoiceSession    voice.VoiceSession
	PlaybackHandler PlaybackHandler
	PlaylistHandler PlaylistHandler
	SongStorage     ports.PlaylistStorage
	StateStorage    ports.PlayerStateStorage
	Messenger       interfaces.DiscordMessenger
	StorageAudio    ports.StorageAudio
	Logger          logging.Logger
}

type GuildPlayer struct {
	playbackHandler PlaybackHandler
	playlistHandler PlaylistHandler
	voiceSession    voice.VoiceSession
	stateStorage    ports.PlayerStateStorage
	eventCh         chan PlayerEvent
	logger          logging.Logger
	running         atomic.Bool
	mu              sync.RWMutex
}

func NewGuildPlayer(cfg Config) *GuildPlayer {
	return &GuildPlayer{
		playbackHandler: cfg.PlaybackHandler,
		playlistHandler: cfg.PlaylistHandler,
		voiceSession:    cfg.VoiceSession,
		stateStorage:    cfg.StateStorage,
		eventCh:         make(chan PlayerEvent, 100),
		logger:          cfg.Logger,
	}
}

func (gp *GuildPlayer) AddSong(ctx context.Context, textChannelID, voiceChannelID *string, playedSong *entity.PlayedSong) error {
	gp.mu.Lock()
	defer gp.mu.Unlock()

	logger := gp.logger.With(
		zap.String("component", "GuildPlayer"),
		zap.String("method", "AddSong"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	if playedSong == nil || playedSong.DiscordSong == nil {
		logger.Error("Intento de agregar canción inválida")
		return errors.New("canción inválida")
	}

	logger = logger.With(
		zap.String("song_id", playedSong.DiscordSong.ID),
		zap.String("title", playedSong.DiscordSong.TitleTrack),
	)

	if err := gp.playlistHandler.AddSong(ctx, playedSong); err != nil {
		logger.Error("Error al agregar canción a playlist",
			zap.Error(err),
			zap.Any("text_channel", textChannelID),
			zap.Any("voice_channel", voiceChannelID))
		return fmt.Errorf("error al agregar la canción: %w", err)
	}

	gp.eventCh <- PlayEvent{
		TextChannelID:  textChannelID,
		VoiceChannelID: voiceChannelID,
	}

	logger.Info("Canción agregada a la lista",
		zap.Any("text_channel", textChannelID),
		zap.Any("voice_channel", voiceChannelID),
	)

	if !gp.running.Load() {
		go func() {
			if err := gp.Run(ctx); err != nil {
				logger.Error("Error al iniciar reproducción", zap.Error(err))
			}
		}()
	}

	return nil
}

func (gp *GuildPlayer) SkipSong(ctx context.Context) {
	logger := gp.logger.With(
		zap.String("component", "GuildPlayer"),
		zap.String("method", "SkipSong"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	currentSong, err := gp.stateStorage.GetCurrentTrack(ctx)
	if err != nil {
		logger.Error("Error al obtener canción actual para skip",
			zap.Error(err))
		return
	}

	if currentSong != nil {
		logger = logger.With(
			zap.String("song_id", currentSong.DiscordSong.ID),
			zap.String("title", currentSong.DiscordSong.TitleTrack),
		)
	}

	gp.playbackHandler.Stop(ctx)
	logger.Info("Canción saltada")
}

func (gp *GuildPlayer) Pause(ctx context.Context) error {
	logger := gp.logger.With(
		zap.String("component", "GuildPlayer"),
		zap.String("method", "Pause"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	currentSong, err := gp.stateStorage.GetCurrentTrack(ctx)
	if err != nil {
		logger.Error("Error al obtener canción actual para pausa",
			zap.Error(err))
		return fmt.Errorf("error al obtener canción actual: %w", err)
	}

	if currentSong != nil {
		logger = logger.With(
			zap.String("song_id", currentSong.DiscordSong.ID),
			zap.String("title", currentSong.DiscordSong.TitleTrack),
		)
	}

	if err := gp.playbackHandler.Pause(ctx); err != nil {
		logger.Error("Error al pausar la reproducción",
			zap.Error(err))
		return err
	}

	logger.Info("Reproducción pausada")
	return nil
}

func (gp *GuildPlayer) Resume(ctx context.Context) error {
	logger := gp.logger.With(
		zap.String("component", "GuildPlayer"),
		zap.String("method", "Resume"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	currentSong, err := gp.stateStorage.GetCurrentTrack(ctx)
	if err != nil {
		logger.Error("Error al obtener canción actual para reanudar",
			zap.Error(err))
		return fmt.Errorf("error al obtener canción actual: %w", err)
	}

	if currentSong != nil {
		logger = logger.With(
			zap.String("song_id", currentSong.DiscordSong.ID),
			zap.String("title", currentSong.DiscordSong.TitleTrack),
		)
	}

	if err := gp.playbackHandler.Resume(ctx); err != nil {
		logger.Error("Error al reanudar la reproducción",
			zap.Error(err))
		return err
	}
	gp.logger.Debug("Reproducción reanudada")
	return nil
}

func (gp *GuildPlayer) Stop(ctx context.Context) error {
	logger := gp.logger.With(
		zap.String("component", "GuildPlayer"),
		zap.String("method", "Stop"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	gp.clearEventChannel()

	currentSong, err := gp.stateStorage.GetCurrentTrack(ctx)
	if err != nil {
		logger.Error("Error al obtener canción actual para detener",
			zap.Error(err))
		return fmt.Errorf("error al obtener canción actual: %w", err)
	}

	if currentSong != nil {
		logger = logger.With(
			zap.String("song_id", currentSong.DiscordSong.ID),
			zap.String("title", currentSong.DiscordSong.TitleTrack),
		)
	}

	if err := gp.playlistHandler.ClearPlaylist(ctx); err != nil {
		logger.Error("Error al limpiar la lista de reproducción",
			zap.Error(err))
		return fmt.Errorf("error al limpiar la lista: %w", err)
	}

	gp.playbackHandler.Stop(ctx)
	if err := gp.voiceSession.LeaveVoiceChannel(ctx); err != nil {
		logger.Error("Error al abandonar el canal de voz",
			zap.Error(err))
		return fmt.Errorf("error al abandonar el canal de voz: %w", err)
	}

	logger.Info("Reproducción detenida y lista limpiada")
	return nil
}

func (gp *GuildPlayer) clearEventChannel() {
	for {
		select {
		case <-gp.eventCh:
		default:
			return
		}
	}
}

func (gp *GuildPlayer) RemoveSong(ctx context.Context, position int) (*entity.DiscordEntity, error) {
	logger := gp.logger.With(
		zap.String("component", "GuildPlayer"),
		zap.String("method", "RemoveSong"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.Int("position", position),
	)

	song, err := gp.playlistHandler.RemoveSong(ctx, position)
	if err != nil {
		logger.Error("Error al remover canción de la playlist",
			zap.Error(err))
		return nil, fmt.Errorf("error al remover la cancion: %w", err)
	}

	logger.Info("Canción removida de la playlist",
		zap.String("song_id", song.ID),
		zap.String("title", song.TitleTrack))

	return song, nil
}

func (gp *GuildPlayer) GetPlaylist(ctx context.Context) ([]string, error) {
	gp.mu.RLock()
	defer gp.mu.RUnlock()

	logger := gp.logger.With(
		zap.String("component", "GuildPlayer"),
		zap.String("method", "GetPlaylist"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	playlist, err := gp.playlistHandler.GetPlaylist(ctx)
	if err != nil {
		logger.Error("Error al obtener la playlist",
			zap.Error(err))
		return nil, fmt.Errorf("error al obtener la playlist: %w", err)
	}

	logger.Debug("Lista de reproducción obtenida",
		zap.Int("count", len(playlist)))
	return playlist, nil
}

func (gp *GuildPlayer) GetPlayedSong(ctx context.Context) (*entity.PlayedSong, error) {
	logger := gp.logger.With(
		zap.String("component", "GuildPlayer"),
		zap.String("method", "GetPlayedSong"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	currentSong, err := gp.stateStorage.GetCurrentTrack(ctx)
	if err != nil {
		logger.Error("Error al obtener canción actual",
			zap.Error(err))
		return nil, fmt.Errorf("error al obtener la cancion actual: %w", err)
	}

	if currentSong != nil {
		logger.Debug("Canción actual obtenida",
			zap.String("song_id", currentSong.DiscordSong.ID),
			zap.String("title", currentSong.DiscordSong.TitleTrack),
			zap.Int64("position", currentSong.Position))
	} else {
		logger.Debug("No hay canción en reproducción")
	}

	return currentSong, nil
}

func (gp *GuildPlayer) Run(ctx context.Context) error {
	logger := gp.logger.With(
		zap.String("component", "GuildPlayer"),
		zap.String("method", "Run"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	if !gp.running.CompareAndSwap(false, true) {
		logger.Warn("Intento de iniciar reproductor ya en ejecución")
		return errors.New("el reproductor ya está ejecutándose")
	}

	runCtx, cancel := context.WithCancel(ctx)

	logger.Debug("Iniciando GuildPlayer")

	defer func() {
		cancel()
		gp.running.Store(false)
		logger.Debug("GuildPlayer detenido")
	}()

	if err := gp.restoreCurrentTrack(ctx); err != nil {
		logger.Warn("No se pudo restaurar la canción actual", zap.Error(err))
	}

	for {
		select {
		case <-runCtx.Done():
			logger.Info("Contexto cancelado - cerrando GuildPlayer")
			return ctx.Err()

		case event := <-gp.eventCh:
			eventType := event.Type()
			logger.Debug("Evento recibido", zap.String("event_type", string(eventType)))

			if eventType == EventTypeStop {
				logger.Debug("Evento de stop recibido, terminando procesamiento")
				return nil
			}

			if err := gp.handleEvent(runCtx, event); err != nil {
				logger.Error("Error al manejar evento del reproductor",
					zap.String("event", string(eventType)),
					zap.Error(err))

				if errors.Is(err, voice.ErrNoVoiceConnection) {
					select {
					case <-time.After(2 * time.Second):
					case <-runCtx.Done():
						return ctx.Err()
					}
				}
			}
		}
	}
}

func (gp *GuildPlayer) restoreCurrentTrack(ctx context.Context) error {
	logger := gp.logger.With(
		zap.String("component", "GuildPlayer"),
		zap.String("method", "restoreCurrentTrack"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	currentSong, err := gp.stateStorage.GetCurrentTrack(ctx)
	if err != nil {
		logger.Error("Error al obtener estado de canción actual", zap.Error(err))
		return err
	}

	if currentSong == nil {
		logger.Debug("No hay canción actual para restaurar")
		return nil
	}

	currentSong.StartPosition += currentSong.Position

	if err := gp.playlistHandler.AddSong(ctx, currentSong); err != nil {
		logger.Error("Error al restaurar canción actual a la lista",
			zap.Error(err),
			zap.String("song_id", currentSong.DiscordSong.ID))
		return err
	}

	logger.Debug("Canción actual restaurada a la lista",
		zap.String("song_id", currentSong.DiscordSong.ID))
	return nil
}

func (gp *GuildPlayer) handleEvent(ctx context.Context, event PlayerEvent) error {
	logger := gp.logger.With(
		zap.String("event_type", string(event.Type())),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	err := event.HandleEvent(ctx, gp)

	if err != nil {
		logger.Error("Error al procesar evento",
			zap.String("tipo", string(event.Type())),
			zap.Error(err))

		if errors.Is(err, voice.ErrNoVoiceConnection) {
			return err
		}
	} else {
		logger.Debug("Evento procesado exitosamente",
			zap.String("tipo", string(event.Type())))
	}
	return err
}

func (gp *GuildPlayer) handlePlayEvent(ctx context.Context, event PlayEvent) error {
	logger := gp.logger.With(
		zap.String("component", "GuildPlayer"),
		zap.String("method", "handlePlayEvent"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.Any("text_channel", event.TextChannelID),
		zap.Any("voice_channel", event.VoiceChannelID),
	)

	gp.mu.Lock()
	defer gp.mu.Unlock()

	logger.Info("Procesando evento PlayEvent")

	if event.TextChannelID != nil {
		logger.Debug("Estableciendo canal de texto en stateStorage")
		if err := gp.stateStorage.SetTextChannelID(ctx, *event.TextChannelID); err != nil {
			logger.Error("Error al establecer canal de texto",
				zap.Error(err))
			return fmt.Errorf("error al setear el canal de texto: %w", err)
		}
	}

	if event.VoiceChannelID != nil {
		logger.Debug("Estableciendo canal de voz en stateStorage")
		if err := gp.stateStorage.SetVoiceChannelID(ctx, *event.VoiceChannelID); err != nil {
			logger.Error("Error al establecer canal de voz",
				zap.Error(err))
			return fmt.Errorf("error al setear el canal de voz: %w", err)
		}
	}

	voiceChannel, textChannel, err := gp.getVoiceAndTextChannels(ctx)
	if err != nil {
		logger.Error("Error al obtener canales",
			zap.Error(err))
		return fmt.Errorf("error al obtener los canales: %w", err)
	}

	logger = logger.With(
		zap.String("voice_channel", voiceChannel),
		zap.String("text_channel", textChannel),
	)

	logger.Info("Intentando unirse al canal de voz")
	if err := gp.JoinVoiceChannel(ctx, voiceChannel); err != nil {
		logger.Error("Error al unirse al canal de voz",
			zap.Error(err))
		return fmt.Errorf("error al unirse al canal de voz: %w", err)
	}

	logger.Info("Unión al canal de voz exitosa. Iniciando reproducción de playlist...")

	for {
		song, err := gp.playlistHandler.GetNextSong(ctx)
		if errors.Is(err, ErrPlaylistEmpty) {
			logger.Info("Playlist vacía - terminando reproducción")

			select {
			case event := <-gp.eventCh:
				if event.Type() == "play" {
					logger.Debug("Nuevo evento Play recibido - reiniciando reproducción")
					gp.eventCh <- event
					return nil
				}
			default:
				if err := gp.voiceSession.LeaveVoiceChannel(ctx); err != nil {
					logger.Error("Error al salir del canal de voz",
						zap.Error(err))
				}
				logger.Info("Desconectado del canal de voz - playlist vacía")
			}

			return nil
		}
		if err != nil {
			logger.Error("Error al obtener siguiente canción",
				zap.Error(err))
			return fmt.Errorf("error al obtener la siguiente canción: %w", err)
		}

		logger.Info("Reproduciendo canción",
			zap.String("song_id", song.DiscordSong.ID),
			zap.String("title", song.DiscordSong.TitleTrack))

		if err := gp.playbackHandler.Play(ctx, song, textChannel); err != nil {
			if errors.Is(err, context.Canceled) {
				logger.Info("Reproducción cancelada por contexto")
				return nil
			}
			logger.Error("Error en reproducción - reintentando",
				zap.String("song_id", song.DiscordSong.ID),
				zap.Error(err))
			continue
		}

		for gp.playbackHandler.CurrentState() != StateIdle {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func (gp *GuildPlayer) getVoiceAndTextChannels(ctx context.Context) (voiceChannel string, textChannel string, err error) {
	logger := gp.logger.With(
		zap.String("component", "GuildPlayer"),
		zap.String("method", "getVoiceAndTextChannels"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	voiceChannel, err = gp.stateStorage.GetVoiceChannelID(ctx)
	if err != nil {
		logger.Error("Error al obtener canal de voz",
			zap.Error(err))
		return "", "", fmt.Errorf("error al obtener el canal de voz: %w", err)
	}

	textChannel, err = gp.stateStorage.GetTextChannelID(ctx)
	if err != nil {
		logger.Error("Error al obtener canal de texto",
			zap.Error(err))
		return "", "", fmt.Errorf("error al obtener el canal de texto: %w", err)
	}

	logger.Debug("Canales obtenidos",
		zap.String("voice_channel", voiceChannel),
		zap.String("text_channel", textChannel))

	return voiceChannel, textChannel, nil
}

func (gp *GuildPlayer) JoinVoiceChannel(ctx context.Context, channelID string) error {
	voiceSession := gp.voiceSession
	return voiceSession.JoinVoiceChannel(ctx, channelID)
}

func (gp *GuildPlayer) StateStorage() ports.PlayerStateStorage {
	return gp.stateStorage
}

func (gp *GuildPlayer) Close() error {
	logger := gp.logger.With(
		zap.String("component", "GuildPlayer"),
		zap.String("method", "Close"),
	)

	logger.Info("Cerrando recursos del GuildPlayer")

	stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := gp.Stop(stopCtx); err != nil {
		logger.Error("Error al detener reproducción durante cierre", zap.Error(err))
	}

	gp.clearEventChannel()

	return nil
}
