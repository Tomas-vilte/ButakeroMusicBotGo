//go:build !integration

package player

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"testing"
	"time"
)

func TestPlaybackController_Play(t *testing.T) {
	t.Run("debería comenzar la reproducción cuando no hay nada en reproducción", func(t *testing.T) {
		// arrange
		playbackStarted := make(chan struct{})
		mockVoiceSession := new(MockVoiceSession)
		mockStorageAudio := new(MockStorageAudio)
		mockStateStorage := new(MockPlayerStateStorage)
		mockMessenger := new(MockDiscordMessenger)
		logger := new(logging.MockLogger)

		pc := NewPlaybackController(
			mockVoiceSession,
			mockStorageAudio,
			mockStateStorage,
			mockMessenger,
			logger,
		)

		song := &entity.PlayedSong{
			DiscordSong: &entity.DiscordEntity{
				ID:         "123",
				TitleTrack: "Test Song",
				FilePath:   "test.mp3",
				DurationMs: 300000,
			},
		}

		logger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return(logger).Once()

		mockStateStorage.On("SetCurrentTrack", mock.Anything, song).Return(nil).Once()
		mockStorageAudio.On("GetAudio", mock.Anything, "test.mp3").Return(&mockReadCloser{}, nil).Once()
		mockMessenger.On("SendPlayStatus", "text-channel", song).Return("msg123", nil).Once()

		mockVoiceSession.On("SendAudio", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			close(playbackStarted)
			select {}
		}).Once()

		mockMessenger.On("UpdatePlayStatus", "text-channel", "msg123").Return("msg123", nil).Once()
		mockStateStorage.On("SetCurrentTrack", mock.Anything, (*entity.PlayedSong)(nil)).Return(nil).Maybe()

		// act
		err := pc.Play(context.Background(), song, "text-channel")

		<-playbackStarted

		// assert
		assert.NoError(t, err)
		assert.Equal(t, StatePlaying, pc.CurrentState(), "El estado debería ser 'playing' durante la reproducción")

		mockStateStorage.AssertCalled(t, "SetCurrentTrack", mock.Anything, song)
		mockStorageAudio.AssertCalled(t, "GetAudio", mock.Anything, "test.mp3")
		mockMessenger.AssertCalled(t, "SendPlayStatus", "text-channel", song)
		mockVoiceSession.AssertCalled(t, "SendAudio", mock.Anything, mock.Anything)

		pc.Stop(context.Background())
		assert.Equal(t, StateIdle, pc.CurrentState(), "El estado debería ser 'idle' después de Stop()")

	})

	t.Run("debería devolver error cuando ya se está reproduciendo", func(t *testing.T) {
		mockVoiceSession := new(MockVoiceSession)
		mockStorageAudio := new(MockStorageAudio)
		mockStateStorage := new(MockPlayerStateStorage)
		mockMessenger := new(MockDiscordMessenger)
		logger := new(logging.MockLogger)

		logger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return(logger)

		pc := NewPlaybackController(mockVoiceSession, mockStorageAudio, mockStateStorage, mockMessenger, logger)
		pc.stateManager.SetState(StatePlaying)

		song := &entity.PlayedSong{
			DiscordSong: &entity.DiscordEntity{
				ID: "123",
			},
		}

		err := pc.Play(context.Background(), song, "text-channel")
		assert.Error(t, err)
		assert.Equal(t, "ya se está reproduciendo", err.Error())
	})

	t.Run("debería pausar la reproducción correctamente", func(t *testing.T) {
		mockVoiceSession := new(MockVoiceSession)
		mockStorageAudio := new(MockStorageAudio)
		mockStateStorage := new(MockPlayerStateStorage)
		mockMessenger := new(MockDiscordMessenger)
		logger := new(logging.MockLogger)

		logger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return(logger)

		pc := NewPlaybackController(mockVoiceSession, mockStorageAudio, mockStateStorage, mockMessenger, logger)
		pc.stateManager.SetState(StatePlaying)
		pc.currentSong = &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{ID: "123"}}

		mockVoiceSession.On("Pause").Once()

		err := pc.Pause(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, StatePaused, pc.CurrentState())
		mockVoiceSession.AssertExpectations(t)
	})

	t.Run("debería devolver error al pausar cuando no se está reproduciendo", func(t *testing.T) {
		mockVoiceSession := new(MockVoiceSession)
		mockStorageAudio := new(MockStorageAudio)
		mockStateStorage := new(MockPlayerStateStorage)
		mockMessenger := new(MockDiscordMessenger)
		logger := new(logging.MockLogger)

		logger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return(logger)

		pc := NewPlaybackController(mockVoiceSession, mockStorageAudio, mockStateStorage, mockMessenger, logger)

		err := pc.Pause(context.Background())
		assert.Error(t, err)
		assert.Equal(t, "no se está reproduciendo", err.Error())
	})

	t.Run("debería reanudar la reproducción correctamente", func(t *testing.T) {
		mockVoiceSession := new(MockVoiceSession)
		mockStorageAudio := new(MockStorageAudio)
		mockStateStorage := new(MockPlayerStateStorage)
		mockMessenger := new(MockDiscordMessenger)
		logger := new(logging.MockLogger)

		logger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return(logger)

		pc := NewPlaybackController(mockVoiceSession, mockStorageAudio, mockStateStorage, mockMessenger, logger)
		pc.stateManager.SetState(StatePaused)
		pc.currentSong = &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{ID: "123"}}
		pc.isPaused.Store(true)

		mockVoiceSession.On("Resume").Once()

		err := pc.Resume(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, StatePlaying, pc.CurrentState())
		assert.False(t, pc.isPaused.Load())
		mockVoiceSession.AssertExpectations(t)
	})

	t.Run("debería devolver error al reanudar cuando no está pausado", func(t *testing.T) {
		mockVoiceSession := new(MockVoiceSession)
		mockStorageAudio := new(MockStorageAudio)
		mockStateStorage := new(MockPlayerStateStorage)
		mockMessenger := new(MockDiscordMessenger)
		logger := new(logging.MockLogger)

		pc := NewPlaybackController(mockVoiceSession, mockStorageAudio, mockStateStorage, mockMessenger, logger)

		logger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Warn", mock.Anything, mock.Anything, mock.Anything).Return(logger)

		err := pc.Resume(context.Background())
		assert.Error(t, err)
		assert.Equal(t, "no se está pausando", err.Error())
	})

	t.Run("debería detener la reproducción correctamente", func(t *testing.T) {
		mockVoiceSession := new(MockVoiceSession)
		mockStorageAudio := new(MockStorageAudio)
		mockStateStorage := new(MockPlayerStateStorage)
		mockMessenger := new(MockDiscordMessenger)
		logger := new(logging.MockLogger)

		pc := NewPlaybackController(mockVoiceSession, mockStorageAudio, mockStateStorage, mockMessenger, logger)

		pc.stateManager.SetState(StatePlaying)
		pc.currentSong = &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{ID: "123"}}
		pc.currentCancel = func() {} // Mock cancel function

		logger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return(logger)

		pc.Stop(context.Background())

		assert.Equal(t, StateIdle, pc.CurrentState())
		assert.Nil(t, pc.currentSong)
		assert.Nil(t, pc.currentCancel)
	})

	t.Run("debería manejar error al obtener audio", func(t *testing.T) {
		mockVoiceSession := new(MockVoiceSession)
		mockStorageAudio := new(MockStorageAudio)
		mockStateStorage := new(MockPlayerStateStorage)
		mockMessenger := new(MockDiscordMessenger)
		logger := new(logging.MockLogger)

		pc := NewPlaybackController(mockVoiceSession, mockStorageAudio, mockStateStorage, mockMessenger, logger)
		pc.stateManager.SetState(StatePlaying)

		song := &entity.PlayedSong{
			DiscordSong: &entity.DiscordEntity{
				ID:       "123",
				FilePath: "test.mp3",
			},
		}

		logger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		mockStateStorage.On("SetCurrentTrack", mock.Anything, song).Return(nil).Once()
		mockStateStorage.On("SetCurrentTrack", mock.Anything, (*entity.PlayedSong)(nil)).Return(nil).Once()
		mockMessenger.On("SendPlayStatus", "text-channel", song).Return("msg123", nil).Once()
		mockStorageAudio.On("GetAudio", mock.Anything, "test.mp3").Return(nil, errors.New("audio error")).Once()

		pc.playSong(context.Background(), song, "text-channel")

		mockStateStorage.AssertExpectations(t)
		mockMessenger.AssertExpectations(t)
		mockStorageAudio.AssertExpectations(t)
	})

	t.Run("debería manejar error al enviar estado de reproducción", func(t *testing.T) {
		mockVoiceSession := new(MockVoiceSession)
		mockStorageAudio := new(MockStorageAudio)
		mockStateStorage := new(MockPlayerStateStorage)
		mockMessenger := new(MockDiscordMessenger)
		logger := new(logging.MockLogger)

		pc := NewPlaybackController(mockVoiceSession, mockStorageAudio, mockStateStorage, mockMessenger, logger)
		pc.stateManager.SetState(StatePlaying)

		song := &entity.PlayedSong{
			DiscordSong: &entity.DiscordEntity{
				ID:       "123",
				FilePath: "test.mp3",
			},
		}

		logger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		mockStateStorage.On("SetCurrentTrack", mock.Anything, song).Return(nil).Once()
		mockMessenger.On("SendPlayStatus", "text-channel", song).Return("", errors.New("send error")).Once()
		mockStorageAudio.On("GetAudio", mock.Anything, "test.mp3").Return(nil, errors.New("audio error")).Once()

		pc.playSong(context.Background(), song, "text-channel")

		mockStateStorage.AssertExpectations(t)
		mockMessenger.AssertExpectations(t)
	})

	t.Run("debería manejar error al enviar audio pero ignorar context.Canceled", func(t *testing.T) {
		// arrange
		mockVoiceSession := new(MockVoiceSession)
		mockStorageAudio := new(MockStorageAudio)
		mockStateStorage := new(MockPlayerStateStorage)
		mockMessenger := new(MockDiscordMessenger)
		logger := new(logging.MockLogger)

		pc := NewPlaybackController(
			mockVoiceSession,
			mockStorageAudio,
			mockStateStorage,
			mockMessenger,
			logger,
		)

		song := &entity.PlayedSong{
			DiscordSong: &entity.DiscordEntity{
				ID:         "123",
				TitleTrack: "Test Song",
				FilePath:   "test.mp3",
				DurationMs: 300000,
			},
		}

		logger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Error", "Error al reproducir audio", mock.Anything).Return(logger).Once()

		mockStateStorage.On("SetCurrentTrack", mock.Anything, song).Return(nil).Once()
		mockStorageAudio.On("GetAudio", mock.Anything, "test.mp3").Return(&mockReadCloser{}, nil).Once()
		mockMessenger.On("SendPlayStatus", "text-channel", song).Return("msg123", nil).Once()

		mockVoiceSession.On("SendAudio", mock.Anything, mock.Anything).Return(errors.New("audio send error")).Once()

		mockMessenger.On("UpdatePlayStatus", "text-channel", "msg123", mock.Anything).Return(nil).Once()
		mockStateStorage.On("SetCurrentTrack", mock.Anything, (*entity.PlayedSong)(nil)).Return(nil).Once()

		// act
		err := pc.Play(context.Background(), song, "text-channel")
		if err != nil {
			return
		}

		time.Sleep(100 * time.Millisecond)

		// assert
		logger.AssertCalled(t, "Error", "Error al reproducir audio", mock.Anything)
		assert.Equal(t, StateIdle, pc.CurrentState())
		mockVoiceSession.AssertExpectations(t)
		mockStateStorage.AssertExpectations(t)
		mockMessenger.AssertExpectations(t)
	})

	t.Run("debería manejar error al actualizar estado final", func(t *testing.T) {
		// arrange
		mockVoiceSession := new(MockVoiceSession)
		mockStorageAudio := new(MockStorageAudio)
		mockStateStorage := new(MockPlayerStateStorage)
		mockMessenger := new(MockDiscordMessenger)
		logger := new(logging.MockLogger)

		pc := NewPlaybackController(
			mockVoiceSession,
			mockStorageAudio,
			mockStateStorage,
			mockMessenger,
			logger,
		)

		song := &entity.PlayedSong{
			DiscordSong: &entity.DiscordEntity{
				ID:         "123",
				TitleTrack: "Test Song",
				FilePath:   "test.mp3",
				DurationMs: 300000,
			},
		}

		logger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Error", "Error al actualizar estado final", mock.Anything).Return(logger).Once()

		mockStateStorage.On("SetCurrentTrack", mock.Anything, song).Return(nil).Once()
		mockStorageAudio.On("GetAudio", mock.Anything, "test.mp3").Return(&mockReadCloser{}, nil).Once()
		mockMessenger.On("SendPlayStatus", "text-channel", song).Return("msg123", nil).Once()

		mockVoiceSession.On("SendAudio", mock.Anything, mock.Anything).Return(nil).Once()

		mockMessenger.On("UpdatePlayStatus", "text-channel", "msg123", mock.Anything).Return(errors.New("update error")).Once()

		mockStateStorage.On("SetCurrentTrack", mock.Anything, (*entity.PlayedSong)(nil)).Return(nil).Once()

		// act
		err := pc.Play(context.Background(), song, "text-channel")
		if err != nil {
			return
		}
		time.Sleep(100 * time.Millisecond)

		// assert
		logger.AssertCalled(t, "Error", "Error al actualizar estado final", mock.Anything)
		assert.Equal(t, StateIdle, pc.CurrentState())
		mockMessenger.AssertCalled(t, "UpdatePlayStatus", "text-channel", "msg123", mock.Anything)
	})

	t.Run("debería manejar error al limpiar estado de canción actual", func(t *testing.T) {
		// arrange
		mockVoiceSession := new(MockVoiceSession)
		mockStorageAudio := new(MockStorageAudio)
		mockStateStorage := new(MockPlayerStateStorage)
		mockMessenger := new(MockDiscordMessenger)
		logger := new(logging.MockLogger)

		pc := NewPlaybackController(
			mockVoiceSession,
			mockStorageAudio,
			mockStateStorage,
			mockMessenger,
			logger,
		)

		song := &entity.PlayedSong{
			DiscordSong: &entity.DiscordEntity{
				ID:         "123",
				TitleTrack: "Test Song",
				FilePath:   "test.mp3",
				DurationMs: 300000,
			},
		}

		logger.On("With", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Info", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return(logger)
		logger.On("Error", "Error al limpiar estado de canción actual", mock.Anything).Return(logger).Once()

		mockStateStorage.On("SetCurrentTrack", mock.Anything, song).Return(nil).Once()
		mockStorageAudio.On("GetAudio", mock.Anything, "test.mp3").Return(&mockReadCloser{}, nil).Once()
		mockMessenger.On("SendPlayStatus", "text-channel", song).Return("msg123", nil).Once()

		mockVoiceSession.On("SendAudio", mock.Anything, mock.Anything).Return(nil).Once()

		mockMessenger.On("UpdatePlayStatus", "text-channel", "msg123", mock.Anything).Return(nil).Once()

		mockStateStorage.On("SetCurrentTrack", mock.Anything, (*entity.PlayedSong)(nil)).Return(errors.New("clear error")).Once()

		// act
		err := pc.Play(context.Background(), song, "text-channel")
		if err != nil {
			return
		}
		time.Sleep(100 * time.Millisecond)

		// Assert
		logger.AssertCalled(t, "Error", "Error al limpiar estado de canción actual", mock.Anything)
		assert.Equal(t, StateIdle, pc.CurrentState())
		mockStateStorage.AssertCalled(t, "SetCurrentTrack", mock.Anything, (*entity.PlayedSong)(nil))
	})
}

type mockReadCloser struct{}

func (m *mockReadCloser) Read(_ []byte) (n int, err error) {
	return 0, io.EOF
}

func (m *mockReadCloser) Close() error {
	return nil
}
