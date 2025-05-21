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

func (m *MockGuildPlayer) SkipSong(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockGuildPlayer) AddSong(ctx context.Context, textChannelID, voiceChannelID *string, playedSong *entity.PlayedSong) error {
	args := m.Called(ctx, textChannelID, voiceChannelID, playedSong)
	return args.Error(0)
}

func (m *MockGuildPlayer) RemoveSong(ctx context.Context, position int) (*entity.PlayedSong, error) {
	args := m.Called(ctx, position)
	return args.Get(0).(*entity.PlayedSong), args.Error(1)
}

func (m *MockGuildPlayer) GetPlaylist(ctx context.Context) ([]*entity.PlayedSong, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.PlayedSong), args.Error(1)
}

func (m *MockGuildPlayer) GetPlayedSong(ctx context.Context) (*entity.PlayedSong, error) {
	args := m.Called(ctx)
	return args.Get(0).(*entity.PlayedSong), args.Error(1)
}

func (m *MockGuildPlayer) MoveToVoiceChannel(ctx context.Context, newChannelID string) error {
	args := m.Called(ctx, newChannelID)
	return args.Error(0)
}

func (m *MockGuildPlayer) Close() error {
	args := m.Called()
	return args.Error(0)
}
