package queuing

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

type MockDiscordGoClient struct {
	mock.Mock
}

func (m *MockDiscordGoClient) SendMessage(channelID string, embed *discordgo.MessageEmbed) error {
	args := m.Called(channelID, embed)
	return args.Error(0)
}

func (m *MockDiscordGoClient) SendMessageToServers(embed *discordgo.MessageEmbed) error {
	args := m.Called(embed)
	return args.Error(0)
}

// MockLogger Mock para la interfaz Logger
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}
