package discord

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/stretchr/testify/mock"
)

type MockPlayerFactory struct {
	mock.Mock
}

func (m *MockPlayerFactory) CreatePlayer(guildID string) (ports.GuildPlayer, error) {
	args := m.Called(guildID)
	return args.Get(0).(ports.GuildPlayer), args.Error(1)
}

type MockGuildPlayer struct {
	mock.Mock
}

func (m *MockGuildPlayer) Run(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockGuildPlayer) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockGuildPlayer) Pause() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockGuildPlayer) Resume() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockGuildPlayer) SkipSong() {
	m.Called()
}

func (m *MockGuildPlayer) AddSong(textChannelID, voiceChannelID *string, playedSong *entity.PlayedSong) error {
	args := m.Called(textChannelID, voiceChannelID, playedSong)
	return args.Error(0)
}

func (m *MockGuildPlayer) RemoveSong(position int) (*entity.DiscordEntity, error) {
	args := m.Called(position)
	return args.Get(0).(*entity.DiscordEntity), args.Error(1)
}

func (m *MockGuildPlayer) GetPlaylist() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockGuildPlayer) GetPlayedSong() (*entity.PlayedSong, error) {
	args := m.Called()
	return args.Get(0).(*entity.PlayedSong), args.Error(1)
}

func (m *MockGuildPlayer) StateStorage() ports.StateStorage {
	args := m.Called()
	return args.Get(0).(ports.StateStorage)
}

func (m *MockGuildPlayer) JoinVoiceChannel(channelID string) error {
	args := m.Called(channelID)
	return args.Error(0)
}

type MockStateStorage struct {
	mock.Mock
}

func (m *MockStateStorage) SetVoiceChannel(channelID string) error {
	args := m.Called(channelID)
	return args.Error(0)
}

func (m *MockStateStorage) GetVoiceChannel() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockStateStorage) SetTextChannel(channelID string) error {
	args := m.Called(channelID)
	return args.Error(0)
}

func (m *MockStateStorage) GetTextChannel() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockStateStorage) SetCurrentSong(song *entity.PlayedSong) error {
	args := m.Called(song)
	return args.Error(0)
}

func (m *MockStateStorage) GetCurrentSong() (*entity.PlayedSong, error) {
	args := m.Called()
	return args.Get(0).(*entity.PlayedSong), args.Error(1)
}
