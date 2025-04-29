package player

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/interfaces"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/voice"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
)

var (
	ErrAlreadyPlaying = errors.New("ya se está reproduciendo")
	ErrNotPlaying     = errors.New("no se está reproduciendo")
	ErrNotPaused      = errors.New("no se está pausando")
)

type PlaybackController struct {
	voiceSession  voice.VoiceSession
	storageAudio  ports.StorageAudio
	stateStorage  ports.PlayerStateStorage
	messenger     interfaces.DiscordMessenger
	logger        logging.Logger
	stateManager  *StateManager
	currentCancel context.CancelFunc
	isPaused      atomic.Bool
	mu            sync.RWMutex
	currentSong   *entity.PlayedSong
	playMsgID     string
}

func NewPlaybackController(
	voiceSession voice.VoiceSession,
	storageAudio ports.StorageAudio,
	stateStorage ports.PlayerStateStorage,
	messenger interfaces.DiscordMessenger,
	logger logging.Logger,
) *PlaybackController {
	return &PlaybackController{
		voiceSession: voiceSession,
		storageAudio: storageAudio,
		stateStorage: stateStorage,
		messenger:    messenger,
		logger:       logger,
		stateManager: NewStateManager(),
	}
}

func (pc *PlaybackController) Play(ctx context.Context, song *entity.PlayedSong, textChannel string) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	logger := pc.getLogger(ctx, "Play", song.DiscordSong.ID)

	if pc.stateManager.GetState() == StatePlaying {
		logger.Warn("Intento de reproducción cuando ya se está reproduciendo")
		return ErrAlreadyPlaying
	}

	songCtx, cancel := context.WithCancel(ctx)
	pc.currentCancel = cancel
	pc.currentSong = song
	pc.stateManager.SetState(StatePlaying)

	logger.Info("Iniciando reproducción",
		zap.String("title", song.DiscordSong.TitleTrack),
		zap.String("channel", textChannel),
	)
	go pc.playSong(songCtx, song, textChannel)
	return nil
}

func (pc *PlaybackController) Pause(ctx context.Context) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	logger := pc.getLogger(ctx, "Pause", "")

	if pc.stateManager.GetState() != StatePlaying {
		logger.Warn("Intento de pausa cuando no se está reproduciendo")
		return ErrNotPlaying
	}

	pc.stateManager.SetState(StatePaused)
	pc.isPaused.Store(true)
	pc.voiceSession.Pause()

	if pc.currentSong != nil {
		logger.Info("Reproducción pausada", zap.String("song_id", pc.currentSong.DiscordSong.ID))
	}
	return nil
}

func (pc *PlaybackController) Resume(ctx context.Context) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	logger := pc.getLogger(ctx, "Resume", "")

	if pc.stateManager.GetState() != StatePaused {
		logger.Warn("Intento de reanudación cuando no se está pausando")
		return ErrNotPaused
	}

	pc.stateManager.SetState(StatePlaying)
	pc.isPaused.Store(false)
	pc.voiceSession.Resume()

	if pc.currentSong != nil {
		logger.Info("Reproducción reanudada", zap.String("song_id", pc.currentSong.DiscordSong.ID))
	}
	return nil
}

func (pc *PlaybackController) Stop(ctx context.Context) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	logger := pc.getLogger(ctx, "Stop", "")

	if pc.currentCancel != nil {
		pc.currentCancel()
		pc.currentCancel = nil
	}
	pc.stateManager.SetState(StateIdle)

	if pc.currentSong != nil {
		logger.Info("Reproducción detenida", zap.String("song_id", pc.currentSong.DiscordSong.ID))
		pc.currentSong = nil
	}
}

func (pc *PlaybackController) CurrentState() PlayerState {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	return pc.stateManager.GetState()
}

func (pc *PlaybackController) getLogger(ctx context.Context, method, songID string) logging.Logger {
	fields := []zap.Field{
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("component", "PlaybackController"),
		zap.String("method", method),
	}

	if songID != "" {
		fields = append(fields, zap.String("song_id", songID))
	}

	return pc.logger.With(fields...)
}

func (pc *PlaybackController) playSong(ctx context.Context, song *entity.PlayedSong, textChannel string) {
	logger := pc.getLogger(ctx, "playSong", song.DiscordSong.ID)

	if err := pc.setInitialPlaybackState(ctx, song, textChannel); err != nil {
		logger.Error("Error en la configuración inicial de reproducción", zap.Error(err))
		return
	}

	audioData, err := pc.storageAudio.GetAudio(ctx, song.DiscordSong.FilePath)
	if err != nil {
		logger.Error("Error al obtener audio", zap.Error(err))
		pc.cleanupAfterPlayback(ctx)
		return
	}

	defer func() {
		if err := audioData.Close(); err != nil {
			logger.Error("Error al cerrar audio", zap.Error(err))
		}
	}()

	done := pc.startPlaybackMonitoring(ctx, song, textChannel)
	defer close(done)

	if err := pc.streamAudio(ctx, audioData); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("Error al reproducir audio", zap.Error(err))
	}

	pc.finalizePlayback(ctx, song, textChannel)
	logger.Info("Reproducción completada")
}

func (pc *PlaybackController) setInitialPlaybackState(ctx context.Context, song *entity.PlayedSong, textChannel string) error {
	logger := pc.getLogger(ctx, "setInitialPlaybackState", song.DiscordSong.ID)

	if err := pc.stateStorage.SetCurrentTrack(ctx, song); err != nil {
		logger.Error("Error al establecer canción actual", zap.Error(err))
		return err
	}

	msgID, err := pc.messenger.SendPlayStatus(textChannel, song)
	if err != nil {
		logger.Error("Error al enviar estado de reproducción", zap.Error(err))
		return err
	}

	pc.playMsgID = msgID
	logger.Debug("Mensaje de estado enviado", zap.String("message_id", msgID))
	return nil
}

func (pc *PlaybackController) startPlaybackMonitoring(ctx context.Context, song *entity.PlayedSong, textChannel string) chan struct{} {
	logger := pc.getLogger(ctx, "startPlaybackMonitoring", song.DiscordSong.ID)
	ticker := time.NewTicker(1 * time.Second)
	done := make(chan struct{})

	startTime := time.Now()
	var pauseStart time.Time
	var totalPaused time.Duration

	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				pc.mu.Lock()
				if pc.currentSong != nil {
					if pc.isPaused.Load() {
						if pauseStart.IsZero() {
							pauseStart = time.Now()
						}
					} else {
						if !pauseStart.IsZero() {
							totalPaused += time.Since(pauseStart)
							pauseStart = time.Time{}
						}
						pc.currentSong.Position = time.Since(startTime).Milliseconds() - totalPaused.Milliseconds()
						if err := pc.messenger.UpdatePlayStatus(textChannel, pc.playMsgID, pc.currentSong); err != nil {
							logger.Error("Error al actualizar estado", zap.Error(err))
						}
					}
				}
				pc.mu.Unlock()
			case <-ctx.Done():
				return
			case <-done:
				return
			}
		}
	}()

	return done
}

func (pc *PlaybackController) streamAudio(ctx context.Context, audioData io.ReadCloser) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if pc.isPaused.Load() {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			return pc.voiceSession.SendAudio(ctx, audioData)
		}
	}
}

func (pc *PlaybackController) finalizePlayback(ctx context.Context, song *entity.PlayedSong, textChannel string) {
	logger := pc.getLogger(ctx, "finalizePlayback", song.DiscordSong.ID)

	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.currentSong != nil {
		pc.currentSong.Position = pc.currentSong.DiscordSong.DurationMs
		if err := pc.messenger.UpdatePlayStatus(textChannel, pc.playMsgID, pc.currentSong); err != nil {
			logger.Error("Error al actualizar estado final", zap.Error(err))
		}
	}

	pc.cleanupAfterPlayback(ctx)
}

func (pc *PlaybackController) cleanupAfterPlayback(ctx context.Context) {
	logger := pc.getLogger(ctx, "cleanupAfterPlayback", "")

	if err := pc.stateStorage.SetCurrentTrack(ctx, nil); err != nil {
		logger.Error("Error al limpiar estado de canción actual", zap.Error(err))
	}

	pc.stateManager.SetState(StateIdle)
	pc.currentSong = nil
}
