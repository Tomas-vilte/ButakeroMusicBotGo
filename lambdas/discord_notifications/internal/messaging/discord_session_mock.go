package messaging

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

// MockDiscordSession es un mock de la interfaz DiscordSession.
type MockDiscordSession struct {
	mock.Mock
}

func (m *MockDiscordSession) ChannelMessageSendEmbed(channelID string, embed *discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	args := m.Called(channelID, embed, options)
	return args.Get(0).(*discordgo.Message), args.Error(1)
}

func (m *MockDiscordSession) UserGuilds(limit int, beforeID, afterID string, withCounts bool, options ...discordgo.RequestOption) ([]*discordgo.UserGuild, error) {
	args := m.Called(limit, beforeID, afterID, withCounts, options)
	return args.Get(0).([]*discordgo.UserGuild), args.Error(1)
}

func (m *MockDiscordSession) GuildChannels(guildID string, options ...discordgo.RequestOption) ([]*discordgo.Channel, error) {
	args := m.Called(guildID, options)
	return args.Get(0).([]*discordgo.Channel), args.Error(1)
}

func (m *MockDiscordSession) GuildChannelCreate(guildID, name string, ctype discordgo.ChannelType, options ...discordgo.RequestOption) (*discordgo.Channel, error) {
	args := m.Called(guildID, name, ctype, options)
	return args.Get(0).(*discordgo.Channel), args.Error(1)
}

func (m *MockDiscordSession) GuildChannelCreateComplex(guildID string, data discordgo.GuildChannelCreateData, options ...discordgo.RequestOption) (*discordgo.Channel, error) {
	args := m.Called(guildID, data, options)
	return args.Get(0).(*discordgo.Channel), args.Error(1)
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
