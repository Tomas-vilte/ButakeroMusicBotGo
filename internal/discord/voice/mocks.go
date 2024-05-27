package voice

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"io"
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
	args := m.Called()
	return args.Get(0).(chan<- []byte)
}

type MockDCAStreamer struct {
	mock.Mock
}

func (m *MockDCAStreamer) StreamDCAData(ctx context.Context, dca io.Reader, opusChan chan<- []byte, positionCallback func(position time.Duration)) error {
	args := m.Called(ctx, dca, opusChan, positionCallback)
	return args.Error(0)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) With(fields ...zap.Field) {
	m.Called(fields)
}
