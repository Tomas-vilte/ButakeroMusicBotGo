package voice

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestChatSessionImpl(t *testing.T) {
	t.Run("Close_Success", func(t *testing.T) {
		// Arrange
		mockDiscordSession := &MockDiscordSessionWrapper{}
		defer mockDiscordSession.AssertExpectations(t)
		mockDiscordSession.On("Close").Return(nil)
		session := &ChatSessionImpl{
			DiscordSession: mockDiscordSession,
		}

		// Act
		err := session.Close()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("JoinVoiceChannel_Success", func(t *testing.T) {
		// Arrange
		mockDiscordSession := &MockDiscordSessionWrapper{}
		defer mockDiscordSession.Mock.AssertExpectations(t)
		mockVoiceConnection := &discordgo.VoiceConnection{}
		mockDiscordSession.On("ChannelVoiceJoin", "test_guild_id", "test_channel_id", false, true).Return(mockVoiceConnection, nil)
		session := &ChatSessionImpl{
			DiscordSession: mockDiscordSession,
			GuildID:        "test_guild_id",
		}

		// Act
		err := session.JoinVoiceChannel("test_channel_id")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, session.voiceConnection)
	})

	t.Run("JoinVoiceChannel_Error", func(t *testing.T) {
		// Arrange
		mockDiscordSession := &MockDiscordSessionWrapper{}
		defer mockDiscordSession.Mock.AssertExpectations(t)
		mockVoiceConnection := &discordgo.VoiceConnection{}
		expectedError := errors.New("error al unirse al canal de voz")
		mockDiscordSession.On("ChannelVoiceJoin", "test_guild_id", "test_channel_id", false, true).Return(mockVoiceConnection, expectedError)
		session := &ChatSessionImpl{
			DiscordSession: mockDiscordSession,
			GuildID:        "test_guild_id",
		}

		// Act
		err := session.JoinVoiceChannel("test_channel_id")

		// Assert
		assert.EqualError(t, err, fmt.Sprintf("mientras se unía al canal de voz: %v", expectedError))
		assert.Nil(t, session.voiceConnection)
	})

	t.Run("LeaveVoiceChannelWithNilConnection", func(t *testing.T) {
		// Arrange
		session := &ChatSessionImpl{}

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
		mockVoiceConnection := &MockVoiceConnectionWrapper{}
		defer mockVoiceConnection.Mock.AssertExpectations(t)
		expectedError := errors.New("error al dejar el canal de voz")
		mockVoiceConnection.On("Disconnect").Return(expectedError)
		session := &ChatSessionImpl{
			voiceConnection: mockVoiceConnection,
		}

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
		mockVoiceConnection := &MockVoiceConnectionWrapper{}
		defer mockVoiceConnection.Mock.AssertExpectations(t)
		mockDCAStreamer := &MockDCAStreamer{}
		defer mockDCAStreamer.Mock.AssertExpectations(t)
		opusSendChan := make(chan<- []byte, 1)
		mockVoiceConnection.On("Speaking", true).Return(nil).Once()
		mockVoiceConnection.On("OpusSendChan").Return(opusSendChan)
		mockDCAStreamer.On("StreamDCAData", mock.Anything, mock.Anything, opusSendChan, mock.Anything).Return(nil)
		mockVoiceConnection.On("Speaking", false).Return(nil).Once()
		session := &ChatSessionImpl{
			voiceConnection: mockVoiceConnection,
			DCAStreamer:     mockDCAStreamer,
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
		mockVoiceConnection := &MockVoiceConnectionWrapper{}
		defer mockVoiceConnection.Mock.AssertExpectations(t)
		mockDCAStreamer := &MockDCAStreamer{}
		defer mockDCAStreamer.Mock.AssertExpectations(t)
		mockVoiceConnection.On("Speaking", true).Return(fmt.Errorf("error al hablar")).Once()
		session := &ChatSessionImpl{
			voiceConnection: mockVoiceConnection,
			DCAStreamer:     mockDCAStreamer,
		}
		audioData := []byte("test_audio_data")
		reader := bytes.NewReader(audioData)
		positionCallback := func(time.Duration) {}

		// Act
		err := session.SendAudio(context.Background(), reader, positionCallback)

		// Assert
		assert.Error(t, err)
	})

	t.Run("SpeakingFalseError", func(t *testing.T) {
		// Arrange
		mockVoiceConnection := &MockVoiceConnectionWrapper{}
		defer mockVoiceConnection.Mock.AssertExpectations(t)
		mockDCAStreamer := &MockDCAStreamer{}
		defer mockDCAStreamer.Mock.AssertExpectations(t)
		opusSendChan := make(chan<- []byte, 1)
		mockVoiceConnection.On("Speaking", true).Return(nil).Once()
		mockVoiceConnection.On("OpusSendChan").Return(opusSendChan)
		mockDCAStreamer.On("StreamDCAData", mock.Anything, mock.Anything, opusSendChan, mock.Anything).Return(nil)
		mockVoiceConnection.On("Speaking", false).Return(fmt.Errorf("error al dejar de hablar")).Once()
		session := &ChatSessionImpl{
			voiceConnection: mockVoiceConnection,
			DCAStreamer:     mockDCAStreamer,
		}
		audioData := []byte("test_audio_data")
		reader := bytes.NewReader(audioData)
		positionCallback := func(time.Duration) {}

		// Act
		err := session.SendAudio(context.Background(), reader, positionCallback)

		// Assert
		assert.Error(t, err)
	})

	t.Run("StreamError", func(t *testing.T) {
		// Arrange
		mockVoiceConnection := new(MockVoiceConnectionWrapper)
		mockDCAStreamer := new(MockDCAStreamer)
		session := &ChatSessionImpl{
			voiceConnection: mockVoiceConnection,
			DCAStreamer:     mockDCAStreamer,
		}
		expectedErr := errors.New("error de transmisión DCA")
		mockVoiceConnection.On("Speaking", true).Return(nil)
		mockDCAStreamer.On("StreamDCAData", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(expectedErr)
		mockVoiceConnection.On("Speaking", false).Return(nil)
		opusSendChan := make(chan<- []byte) // Canal de solo escritura
		mockVoiceConnection.On("OpusSendChan").Return(opusSendChan)

		// Act
		err := session.SendAudio(context.Background(), nil, nil)

		// Assert
		assert.Error(t, err)
		assert.EqualError(t, err, "mientras se transmitían datos DCA: error de transmisión DCA")
	})
}
