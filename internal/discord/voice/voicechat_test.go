package voice

import (
	"bytes"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

type MockDiscordSessionWrapper struct {
	mock.Mock
}

func (m *MockDiscordSessionWrapper) ChannelVoiceJoin(guildID, channelID string, muted, deafened bool) (*discordgo.VoiceConnection, error) {
	args := m.Called(guildID, channelID, muted, deafened)
	return args.Get(0).(*discordgo.VoiceConnection), args.Error(1)
}

func (m *MockDiscordSessionWrapper) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockVoiceConnectionWrapper struct {
	mock.Mock
	opusSendChan chan<- []byte
}

func (m *MockVoiceConnectionWrapper) Disconnect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockVoiceConnectionWrapper) Speaking(flag bool) error {
	args := m.Called(flag)
	return args.Error(0)
}

func (m *MockVoiceConnectionWrapper) OpusSend(data []byte, mode int) (ok bool, err error) {
	args := m.Called(data, mode)
	return args.Bool(0), args.Error(1)
}

func (m *MockVoiceConnectionWrapper) OpusSendChan() chan<- []byte {
	return m.opusSendChan
}

func TestChatSessionImpl_SendAudio(t *testing.T) {
	ctx := context.Background()
	audioData := []byte{0x01, 0x02, 0x03, 0x04}
	reader := bytes.NewReader(audioData)
	positionCallback := func(position time.Duration) {
		// Puedes hacer lo que quieras con la posición aquí, por ejemplo, imprimir
		fmt.Printf("Posición actual: %s\n", position)
	}
	// Creamos un mock para DiscordSessionWrapper
	discordSessionMock := &MockDiscordSessionWrapper{}

	// Creamos un canal para simular el envío de Opus
	bufferSize := 1024 * 1024
	opusSendChan := make(chan []byte, bufferSize) // Define bufferSize según tus necesidades

	// Creamos un mock para VoiceConnectionWrapper
	voiceConnectionMock := &MockVoiceConnectionWrapper{
		opusSendChan: opusSendChan,
	}

	chatSession := &ChatSessionImpl{
		DiscordSession:  discordSessionMock,
		GuildID:         "1231503103279366204",
		voiceConnection: voiceConnectionMock,
	}

	voiceConnectionMock.On("Speaking", true).Return(nil)
	voiceConnectionMock.On("Speaking", false).Return(nil)
	voiceConnectionMock.On("OpusSend", mock.Anything, mock.Anything).Return(true, nil)

	err := chatSession.SendAudio(ctx, reader, positionCallback)

	// Verificamos que no haya errores
	assert.NoError(t, err)
}
