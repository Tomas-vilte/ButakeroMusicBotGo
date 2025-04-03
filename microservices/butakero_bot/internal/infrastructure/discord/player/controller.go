package player

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
)

type PlaybackController struct {
	voiceSession  ports.VoiceSession
	storageAudio  ports.StorageAudio
	stateStorage  ports.StateStorage
	messenger     ports.DiscordMessenger
	logger        logging.Logger
	stateManager  *StateManager
	currentCancel context.CancelFunc
	pauseCh       chan struct{}
	resumeCh      chan struct{}
	mu            sync.Mutex
	currentSong   *entity.PlayedSong
	playMsgID     string
}

func NewPlaybackController(
	voiceSession ports.VoiceSession,
	storageAudio ports.StorageAudio,
	stateStorage ports.StateStorage,
	messenger ports.DiscordMessenger,
	logger logging.Logger,
) *PlaybackController {
	return &PlaybackController{
		voiceSession: voiceSession,
		storageAudio: storageAudio,
		stateStorage: stateStorage,
		messenger:    messenger,
		logger:       logger,
		stateManager: NewStateManager(),
		pauseCh:      make(chan struct{}),
		resumeCh:     make(chan struct{}),
	}
}

func (pc *PlaybackController) Play(ctx context.Context, song *entity.PlayedSong, textChannel string) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.stateManager.GetState() == StatePlaying {
		return errors.New("already playing")
	}

	songCtx, cancel := context.WithCancel(ctx)
	pc.currentCancel = cancel
	pc.currentSong = song
	pc.stateManager.SetState(StatePlaying)

	go pc.playSong(songCtx, song, textChannel)
	return nil
}

func (pc *PlaybackController) Pause() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.stateManager.GetState() != StatePlaying {
		return errors.New("not playing")
	}

	pc.stateManager.SetState(StatePaused)
	pc.voiceSession.Pause()
	return nil
}

func (pc *PlaybackController) Resume() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.stateManager.GetState() != StatePaused {
		return errors.New("not paused")
	}

	pc.stateManager.SetState(StatePlaying)
	pc.voiceSession.Resume()
	return nil
}

func (pc *PlaybackController) Stop() {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.currentCancel != nil {
		pc.currentCancel()
		pc.currentCancel = nil
	}
	pc.stateManager.SetState(StateIdle)
	pc.currentSong = nil
}

func (pc *PlaybackController) CurrentState() PlayerState {
	return pc.stateManager.GetState()
}

func (pc *PlaybackController) GetVoiceSession() ports.VoiceSession {
	return pc.voiceSession
}

func (pc *PlaybackController) CurrentSong() *entity.PlayedSong {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	return pc.currentSong
}

func (pc *PlaybackController) playSong(ctx context.Context, song *entity.PlayedSong, textChannel string) {
	if err := pc.stateStorage.SetCurrentSong(song); err != nil {
		pc.logger.Error("Error al establecer la canción actual", zap.Error(err))
		return
	}

	msgID, err := pc.messenger.SendPlayStatus(textChannel, song)
	if err != nil {
		pc.logger.Error("Error al enviar el estado de reproducción", zap.Error(err))
	} else {
		pc.playMsgID = msgID
	}

	audioData, err := pc.storageAudio.GetAudio(ctx, song.DiscordSong.FilePath)
	if err != nil {
		pc.logger.Error("Error al obtener el audio", zap.Error(err))
		return
	}

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})
	startTime := time.Now()
	pausedTime := time.Duration(0)
	isPaused := false

	go func() {
		defer close(done)
		for {
			select {
			case <-ticker.C:
				pc.mu.Lock()
				if pc.currentSong != nil && !isPaused {
					pc.currentSong.Position = time.Since(startTime).Milliseconds() - pausedTime.Milliseconds()
					if err := pc.messenger.UpdatePlayStatus(textChannel, pc.playMsgID, pc.currentSong); err != nil {
						pc.logger.Error("Error al actualizar el estado de reproducción", zap.Error(err))
					}
				}
				pc.mu.Unlock()
			case <-ctx.Done():
				return
			}
		}
	}()

	pauseStart := time.Time{}
	for {
		select {
		case <-ctx.Done():
			return

		case <-pc.pauseCh:
			if !isPaused {
				isPaused = true
				pauseStart = time.Now()
				pc.logger.Debug("Reproducción pausada")
			}

		case <-pc.resumeCh:
			if isPaused {
				isPaused = false
				pausedTime += time.Since(pauseStart)
				pc.logger.Debug("Reproducción reanudada")
			}

		default:
			if isPaused {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			err := pc.voiceSession.SendAudio(ctx, audioData)
			if err != nil && !errors.Is(err, context.Canceled) {
				pc.logger.Error("Error al reproducir audio", zap.Error(err))
			}

			pc.mu.Lock()
			pc.stateManager.SetState(StateIdle)
			pc.currentSong = nil
			pc.mu.Unlock()

			return
		}
	}
}
