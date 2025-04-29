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

func (m *MockGuildPlayer) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockGuildPlayer) Pause(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockGuildPlayer) Resume(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockGuildPlayer) SkipSong(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockGuildPlayer) AddSong(ctx context.Context, textChannelID, voiceChannelID *string, playedSong *entity.PlayedSong) error {
	args := m.Called(ctx, textChannelID, voiceChannelID, playedSong)
	return args.Error(0)
}

func (m *MockGuildPlayer) RemoveSong(ctx context.Context, position int) (*entity.DiscordEntity, error) {
	args := m.Called(ctx, position)
	return args.Get(0).(*entity.DiscordEntity), args.Error(1)
}

func (m *MockGuildPlayer) GetPlaylist(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockGuildPlayer) GetPlayedSong(ctx context.Context) (*entity.PlayedSong, error) {
	args := m.Called(ctx)
	return args.Get(0).(*entity.PlayedSong), args.Error(1)
}

func (m *MockGuildPlayer) StateStorage() ports.PlayerStateStorage {
	args := m.Called()
	return args.Get(0).(ports.PlayerStateStorage)
}

func (m *MockGuildPlayer) JoinVoiceChannel(ctx context.Context, channelID string) error {
	args := m.Called(ctx, channelID)
	return args.Error(0)
}

func (m *MockGuildPlayer) Close() error {
	args := m.Called()
	return args.Error(0)

}

type MockStateStorage struct {
	mock.Mock
}

func (m *MockStateStorage) SetVoiceChannelID(ctx context.Context, channelID string) error {
	args := m.Called(ctx, channelID)
	return args.Error(0)
}

func (m *MockStateStorage) GetVoiceChannelID(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockStateStorage) SetTextChannelID(ctx context.Context, channelID string) error {
	args := m.Called(ctx, channelID)
	return args.Error(0)
}

func (m *MockStateStorage) GetTextChannelID(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockStateStorage) SetCurrentTrack(ctx context.Context, track *entity.PlayedSong) error {
	args := m.Called(ctx, track)
	return args.Error(0)
}

func (m *MockStateStorage) GetCurrentTrack(ctx context.Context) (*entity.PlayedSong, error) {
	args := m.Called(ctx)
	return args.Get(0).(*entity.PlayedSong), args.Error(1)
}
