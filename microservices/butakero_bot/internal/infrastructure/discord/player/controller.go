package player

import (
	"context"
	"errors"
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

type PlaybackController struct {
	voiceSession  voice.VoiceSession
	storageAudio  ports.StorageAudio
	stateStorage  ports.PlayerStateStorage
	messenger     interfaces.DiscordMessenger
	logger        logging.Logger
	stateManager  *StateManager
	currentCancel context.CancelFunc
	isPaused      atomic.Bool
	mu            sync.Mutex
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

	logger := pc.logger.With(
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("component", "PlaybackController"),
		zap.String("method", "Play"),
		zap.String("song_id", song.DiscordSong.ID),
	)

	if pc.stateManager.GetState() == StatePlaying {
		logger.Warn("Intento de reproducción cuando ya se está reproduciendo")
		return errors.New("already playing")
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

	logger := pc.logger.With(
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("component", "PlaybackController"),
		zap.String("method", "Pause"),
	)

	if pc.stateManager.GetState() != StatePlaying {
		logger.Warn("Intento de pausa cuando no se está reproduciendo")
		return errors.New("not playing")
	}

	pc.stateManager.SetState(StatePaused)
	pc.isPaused.Store(true)
	pc.voiceSession.Pause()

	logger.Info("Reproducción pausada",
		zap.String("song_id", pc.currentSong.DiscordSong.ID),
	)
	return nil
}

func (pc *PlaybackController) Resume(ctx context.Context) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	logger := pc.logger.With(
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("component", "PlaybackController"),
		zap.String("method", "Resume"),
	)

	if pc.stateManager.GetState() != StatePaused {
		logger.Warn("Intento de reanudación cuando no se está pausando")
		return errors.New("not paused")
	}

	pc.stateManager.SetState(StatePlaying)
	pc.isPaused.Store(false)
	pc.voiceSession.Resume()

	logger.Info("Reproducción reanudada",
		zap.String("song_id", pc.currentSong.DiscordSong.ID),
	)
	return nil
}

func (pc *PlaybackController) Stop(ctx context.Context) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	logger := pc.logger.With(
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("component", "PlaybackController"),
		zap.String("method", "Stop"),
	)

	if pc.currentCancel != nil {
		pc.currentCancel()
		pc.currentCancel = nil
	}
	pc.stateManager.SetState(StateIdle)

	if pc.currentSong != nil {
		logger.Info("Reproducción detenida",
			zap.String("song_id", pc.currentSong.DiscordSong.ID))
		pc.currentSong = nil
	}
}

func (pc *PlaybackController) CurrentState() PlayerState {
	return pc.stateManager.GetState()
}

func (pc *PlaybackController) playSong(ctx context.Context, song *entity.PlayedSong, textChannel string) {
	logger := pc.logger.With(
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("component", "PlaybackController"),
		zap.String("method", "playSong"),
		zap.String("song_id", song.DiscordSong.ID),
	)

	if err := pc.stateStorage.SetCurrentTrack(ctx, song); err != nil {
		logger.Error("Error al establecer canción actual", zap.Error(err))
		return
	}

	msgID, err := pc.messenger.SendPlayStatus(textChannel, song)
	if err != nil {
		logger.Error("Error al enviar estado de reproducción", zap.Error(err))
	} else {
		pc.playMsgID = msgID
		logger.Debug("Mensaje de estado enviado", zap.String("message_id", msgID))
	}

	audioData, err := pc.storageAudio.GetAudio(ctx, song.DiscordSong.FilePath)
	if err != nil {
		logger.Error("Error al obtener audio", zap.Error(err))
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})
	startTime := time.Now()
	var pauseStart time.Time
	var totalPaused time.Duration

	go func() {
		defer close(done)
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
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			logger.Debug("Contexto cancelado, terminando reproducción")
			return
		default:
			if pc.isPaused.Load() {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			err := pc.voiceSession.SendAudio(ctx, audioData)
			if err != nil && !errors.Is(err, context.Canceled) {
				logger.Error("Error al enviar audio", zap.Error(err))
			}

			if pc.currentSong != nil {
				pc.currentSong.Position = pc.currentSong.DiscordSong.DurationMs
				if err := pc.messenger.UpdatePlayStatus(textChannel, pc.playMsgID, pc.currentSong); err != nil {
					logger.Error("Error al actualizar estado final", zap.Error(err))
				}
			}

			if err := pc.stateStorage.SetCurrentTrack(ctx, nil); err != nil {
				logger.Error("Error al limpiar estado de canción actual", zap.Error(err))
			}

			pc.mu.Lock()
			pc.stateManager.SetState(StateIdle)
			pc.currentSong = nil
			pc.mu.Unlock()

			logger.Info("Reproducción completada")
			return
		}
	}
}
