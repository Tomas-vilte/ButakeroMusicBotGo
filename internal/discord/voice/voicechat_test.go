package voice

import (
	"bytes"
	"context"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestChatSessionImpl(t *testing.T) {
	t.Run("Close_Success", func(t *testing.T) {
		mockLogger := new(logging.MockLogger)
		// Arrange
		mockDiscordSession := &MockDiscordSessionWrapper{}
		defer mockDiscordSession.AssertExpectations(t)
		mockDiscordSession.On("Close").Return(nil)
		session := NewChatSessionImpl(mockDiscordSession, "", nil, mockLogger)
		mockLogger.On("Info", "Cerrando sesión de Discord...", mock.AnythingOfType("[]zapcore.Field")).Return()

		// Act
		err := session.Close()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("JoinVoiceChannel_Success", func(t *testing.T) {
		mockLogger := new(logging.MockLogger)
		// Arrange
		mockDiscordSession := &MockDiscordSessionWrapper{}
		defer mockDiscordSession.Mock.AssertExpectations(t)
		mockVoiceConnection := &discordgo.VoiceConnection{}
		mockDiscordSession.On("ChannelVoiceJoin", "test_guild_id", "test_channel_id", false, true).Return(mockVoiceConnection, nil)
		session := NewChatSessionImpl(mockDiscordSession, "test_guild_id", nil, mockLogger)

		mockLogger.On("Info", "Uniéndose al canal de voz ...", mock.AnythingOfType("[]zapcore.Field")).Return()
		// Act
		err := session.JoinVoiceChannel("test_channel_id")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, session.voiceConnection)
	})

	t.Run("JoinVoiceChannel_Error", func(t *testing.T) {
		// Arrange
		mockLogger := new(logging.MockLogger)
		mockDiscordSession := &MockDiscordSessionWrapper{}
		defer mockDiscordSession.Mock.AssertExpectations(t)
		expectedError := errors.New("error al unirse al canal de voz")
		mockVoiceConnection := &discordgo.VoiceConnection{}
		mockDiscordSession.On("ChannelVoiceJoin", "test_guild_id", "test_channel_id", false, true).Return(mockVoiceConnection, expectedError)
		session := NewChatSessionImpl(mockDiscordSession, "test_guild_id", nil, mockLogger)

		mockLogger.On("Info", "Uniéndose al canal de voz ...", mock.AnythingOfType("[]zapcore.Field")).Return()
		mockLogger.On("Error", "Error al unirse al canal de voz", mock.AnythingOfType("[]zapcore.Field")).Return()

		// Act
		err := session.JoinVoiceChannel("test_channel_id")

		// Assert
		assert.EqualError(t, err, expectedError.Error())
		assert.Nil(t, session.voiceConnection)
	})

	t.Run("LeaveVoiceChannelWithNilConnection", func(t *testing.T) {
		// Arrange
		mockLogger := new(logging.MockLogger)
		session := NewChatSessionImpl(nil, "", nil, mockLogger)

		// Act
		err := session.LeaveVoiceChannel()

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, session.voiceConnection)
	})

	t.Run("LeaveVoiceChannelSuccess", func(t *testing.T) {
		// Arrange
		mockVoiceConnection := &MockVoiceConnectionWrapper{}
		defer mockVoiceConnection.Mock.AssertExpectations(t)
		mockVoiceConnection.On("Disconnect").Return(nil)
		session := &ChatSessionImpl{
			voiceConnection: mockVoiceConnection,
		}

		// Act
		err := session.LeaveVoiceChannel()

		// Assert
		assert.NoError(t, err)
		assert.Nil(t, session.voiceConnection)
	})

	t.Run("LeaveVoiceChannelWithError", func(t *testing.T) {
		// Arrange
		mockLogger := new(logging.MockLogger)
		mockVoiceConnection := &MockVoiceConnectionWrapper{}
		defer mockVoiceConnection.Mock.AssertExpectations(t)
		expectedError := errors.New("error al dejar el canal de voz")
		mockVoiceConnection.On("Disconnect").Return(expectedError)
		session := &ChatSessionImpl{
			voiceConnection: mockVoiceConnection,
			logger:          mockLogger,
		}
		mockLogger.On("Error", "Error al dejar el canal de voz", mock.AnythingOfType("[]zapcore.Field")).Return()

		// Act
		err := session.LeaveVoiceChannel()

		// Assert
		assert.EqualError(t, err, expectedError.Error())
		assert.Nil(t, session.voiceConnection)
	})
}

func TestChatSessionImpl_SendAudio(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		mockLogger := new(logging.MockLogger)
		mockVoiceConnection := &MockVoiceConnectionWrapper{}
		defer mockVoiceConnection.AssertExpectations(t)
		mockDCAStreamer := &MockDCAStreamer{}
		defer mockDCAStreamer.AssertExpectations(t)

		opusSendChan := make(chan<- []byte, 1)
		mockVoiceConnection.On("Speaking", true).Return(nil).Once()
		mockVoiceConnection.On("OpusSendChan").Return(opusSendChan).Once()
		mockDCAStreamer.On("StreamDCAData", mock.Anything, mock.Anything, opusSendChan, mock.Anything).Return(nil)
		mockVoiceConnection.On("Speaking", false).Return(nil).Once()
		mockLogger.On("Info", "Enviando audio al canal de voz...", mock.AnythingOfType("[]zapcore.Field")).Return()
		session := &ChatSessionImpl{
			voiceConnection: mockVoiceConnection,
			DCAStreamer:     mockDCAStreamer,
			logger:          mockLogger,
		}
		audioData := []byte("test_audio_data")
		reader := bytes.NewReader(audioData)
		positionCallback := func(time.Duration) {}

		// Act
		err := session.SendAudio(context.Background(), reader, positionCallback)

		// Assert
		assert.NoError(t, err)
	})

	t.Run("SpeakingTrueError", func(t *testing.T) {
		// Arrange
		mockLogger := new(logging.MockLogger)
		mockVoiceConnection := &MockVoiceConnectionWrapper{}
		defer mockVoiceConnection.AssertExpectations(t)
		mockDCAStreamer := &MockDCAStreamer{}

		expectedErr := errors.New("error al enviando la canal de voz")
		mockVoiceConnection.On("Speaking", true).Return(expectedErr).Once()
		mockLogger.On("Info", "Enviando audio al canal de voz...", mock.AnythingOfType("[]zapcore.Field")).Return()
		mockLogger.On("Error", "Error al comenzar a hablar: ", mock.AnythingOfType("[]zapcore.Field")).Return()

		session := &ChatSessionImpl{
			voiceConnection: mockVoiceConnection,
			DCAStreamer:     mockDCAStreamer,
			logger:          mockLogger,
		}
		audioData := []byte("test_audio_data")
		reader := bytes.NewReader(audioData)
		positionCallback := func(time.Duration) {}

		// Act
		err := session.SendAudio(context.Background(), reader, positionCallback)

		// Assert
		assert.EqualError(t, err, expectedErr.Error())
	})

	t.Run("SpeakingFalseError", func(t *testing.T) {
		// Arrange
		mockLogger := new(logging.MockLogger)
		mockVoiceConnection := &MockVoiceConnectionWrapper{}
		defer mockVoiceConnection.AssertExpectations(t)
		mockDCAStreamer := &MockDCAStreamer{}
		defer mockDCAStreamer.AssertExpectations(t)

		opusSendChan := make(chan<- []byte, 1)
		expectedErr := errors.New("error al dejar de hablar")
		mockVoiceConnection.On("Speaking", true).Return(nil).Once()
		mockVoiceConnection.On("OpusSendChan").Return(opusSendChan).Once()
		mockDCAStreamer.On("StreamDCAData", mock.Anything, mock.Anything, opusSendChan, mock.Anything).Return(nil)
		mockVoiceConnection.On("Speaking", false).Return(expectedErr).Once()
		mockLogger.On("Info", "Enviando audio al canal de voz...", mock.AnythingOfType("[]zapcore.Field")).Return()
		mockLogger.On("Error", "Error al dejar de hablar: ", mock.AnythingOfType("[]zapcore.Field")).Return()

		session := &ChatSessionImpl{
			voiceConnection: mockVoiceConnection,
			DCAStreamer:     mockDCAStreamer,
			logger:          mockLogger,
		}
		audioData := []byte("test_audio_data")
		reader := bytes.NewReader(audioData)
		positionCallback := func(time.Duration) {}

		// Act
		err := session.SendAudio(context.Background(), reader, positionCallback)

		// Assert
		assert.EqualError(t, err, "error al dejar de hablar")
	})

	t.Run("StreamError", func(t *testing.T) {
		// Arrange
		mockLogger := new(logging.MockLogger)
		mockVoiceConnection := &MockVoiceConnectionWrapper{}
		defer mockVoiceConnection.AssertExpectations(t)
		mockDCAStreamer := &MockDCAStreamer{}
		defer mockDCAStreamer.AssertExpectations(t)

		expectedErr := errors.New("error de transmisión DCA")
		opusSendChan := make(chan<- []byte, 1)
		mockVoiceConnection.On("Speaking", true).Return(nil).Once()
		mockVoiceConnection.On("OpusSendChan").Return(opusSendChan).Once()
		mockDCAStreamer.On("StreamDCAData", mock.Anything, mock.Anything, opusSendChan, mock.Anything).Return(expectedErr)
		mockVoiceConnection.On("Speaking", false).Return(nil).Once() // Ajuste aquí
		mockLogger.On("Info", "Enviando audio al canal de voz...", mock.AnythingOfType("[]zapcore.Field")).Return()
		mockLogger.On("Error", "Error al transmitir datos DCA: ", mock.AnythingOfType("[]zapcore.Field")).Return()

		session := &ChatSessionImpl{
			voiceConnection: mockVoiceConnection,
			DCAStreamer:     mockDCAStreamer,
			logger:          mockLogger,
		}

		// Act
		err := session.SendAudio(context.Background(), nil, nil)

		// Assert
		assert.EqualError(t, err, expectedErr.Error())
	})

	t.Run("OpusSendChanNil", func(t *testing.T) {
		// Arrange
		mockLogger := new(logging.MockLogger)
		mockVoiceConnection := &MockVoiceConnectionWrapper{}
		defer mockVoiceConnection.AssertExpectations(t)
		mockDCAStreamer := &MockDCAStreamer{}
		defer mockDCAStreamer.AssertExpectations(t)

		var opusSendChan chan<- []byte = nil
		mockVoiceConnection.On("Speaking", true).Return(nil).Once()
		mockVoiceConnection.On("OpusSendChan").Return(opusSendChan).Once()
		mockLogger.On("Info", "Enviando audio al canal de voz...", mock.AnythingOfType("[]zapcore.Field")).Return()
		mockLogger.On("Error", "Error canal de envío de Opus no está disponible", mock.AnythingOfType("[]zapcore.Field")).Return()

		session := &ChatSessionImpl{
			voiceConnection: mockVoiceConnection,
			DCAStreamer:     mockDCAStreamer,
			logger:          mockLogger,
		}

		// Act
		err := session.SendAudio(context.Background(), nil, nil)

		// Assert
		assert.EqualError(t, err, "canal de envío de Opus no está disponible")
	})
}
