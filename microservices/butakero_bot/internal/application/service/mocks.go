package service

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model/queue"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/stretchr/testify/mock"
)

type MockSongService struct {
	mock.Mock
}

func (m *MockSongService) GetOrDownloadSong(ctx context.Context, userID, songInput, platform string) (*entity.DiscordEntity, error) {
	args := m.Called(ctx, userID, songInput, platform)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.DiscordEntity), args.Error(1)
}

type MockGuildManager struct {
	mock.Mock
}

func (m *MockGuildManager) CreateGuildPlayer(guildID string) (ports.GuildPlayer, error) {
	args := m.Called(guildID)
	return args.Get(0).(ports.GuildPlayer), args.Error(1)
}

func (m *MockGuildManager) RemoveGuildPlayer(guildID string) error {
	args := m.Called(guildID)
	return args.Error(0)
}

func (m *MockGuildManager) GetGuildPlayer(guildID string) (ports.GuildPlayer, error) {
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

type MockMediaClient struct {
	mock.Mock
}

func (m *MockMediaClient) GetMediaByID(ctx context.Context, videoID string) (*model.Media, error) {
	args := m.Called(ctx, videoID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Media), args.Error(1)
}

func (m *MockMediaClient) SearchMediaByTitle(ctx context.Context, title string) ([]*model.Media, error) {
	args := m.Called(ctx, title)
	return args.Get(0).([]*model.Media), args.Error(1)
}

type MockSongDownloadRequestPublisher struct {
	mock.Mock
}

func (m *MockSongDownloadRequestPublisher) PublishDownloadRequest(ctx context.Context, request *queue.DownloadRequestMessage) error {
	args := m.Called(ctx, request)
	return args.Error(0)
}

func (m *MockSongDownloadRequestPublisher) ClosePublisher() error {
	return m.Called().Error(0)
}

type MockSongDownloadEventSubscriber struct {
	mock.Mock
}

func (m *MockSongDownloadEventSubscriber) SubscribeToDownloadEvents(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockSongDownloadEventSubscriber) DownloadEventsChannel() <-chan *queue.DownloadStatusMessage {
	return m.Called().Get(0).(chan *queue.DownloadStatusMessage)
}

func (m *MockSongDownloadEventSubscriber) CloseSubscription() error {
	return m.Called().Error(0)
}
