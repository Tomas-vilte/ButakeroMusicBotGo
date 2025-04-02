package player

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/decoder"
	"io"
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
	pc.pauseCh <- struct{}{}
	return nil
}

func (pc *PlaybackController) Resume() error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	if pc.stateManager.GetState() != StatePaused {
		return errors.New("not paused")
	}

	pc.stateManager.SetState(StatePlaying)
	pc.resumeCh <- struct{}{}
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

	decoderAudio := decoder.NewBufferedOpusDecoder(audioData)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})
	startTime := time.Now()
	pausedTime := time.Duration(0)
	isPaused := false

	go func() {
		for {
			select {
			case <-ticker.C:
				pc.mu.Lock()
				if pc.currentSong != nil {
					if !isPaused {
						pc.currentSong.Position = time.Since(startTime).Milliseconds() - pausedTime.Milliseconds()
					}
					if err := pc.messenger.UpdatePlayStatus(textChannel, pc.playMsgID, pc.currentSong); err != nil {
						pc.logger.Error("Error al actualizar el estado de reproducción", zap.Error(err))
					}
				}
				pc.mu.Unlock()
			case <-done:
				return
			}
		}
	}()

	if err := pc.voiceSession.GetVoiceConnection().Speaking(true); err != nil {
		pc.logger.Error("Error al iniciar la reproducción", zap.Error(err))
		return
	}
	defer func() {
		if err := pc.voiceSession.GetVoiceConnection().Speaking(false); err != nil {
			pc.logger.Error("Error al detener la reproducción", zap.Error(err))
		}
	}()

	pauseStart := time.Time{}

	for {
		select {
		case <-ctx.Done():
			close(done)
			return

		case <-pc.pauseCh:
			isPaused = true
			pauseStart = time.Now()
			pc.logger.Debug("Reproducción pausada")

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

			frame, err := decoderAudio.OpusFrame()
			if err != nil {
				if errors.Is(err, io.EOF) {
					pc.logger.Debug("Fin del archivo de audio")
					close(done)
					return
				}
				pc.logger.Error("Error al decodificar frame de audio", zap.Error(err))
				close(done)
				return
			}

			select {
			case pc.voiceSession.GetVoiceConnection().OpusSend <- frame:
			case <-time.After(time.Second):
				pc.logger.Error("Tiempo de espera agotado al enviar frame de audio")
				close(done)
				return
			case <-ctx.Done():
				close(done)
				return
			}

			time.Sleep(20 * time.Millisecond)
		}
	}
}
