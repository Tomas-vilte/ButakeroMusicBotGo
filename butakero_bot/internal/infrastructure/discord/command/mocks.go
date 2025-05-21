package command

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/mock"
)

type MockInteractionStorage struct {
	mock.Mock
}

func (m *MockInteractionStorage) SaveSongList(channelID string, list []*entity.DiscordEntity) {
	args := m.Called(channelID, list)
	if args.Get(0) != nil {
		panic(args.Get(0))
	}
}

func (m *MockInteractionStorage) GetSongList(channelID string) []*entity.DiscordEntity {
	args := m.Called(channelID)
	if args.Get(0) != nil {
		panic(args.Get(0))
	}
	return args.Get(0).([]*entity.DiscordEntity)
}

func (m *MockInteractionStorage) DeleteSongList(channelID string) {
	m.Called(channelID)
}

type MockDiscordMessenger struct {
	mock.Mock
}

func (m *MockDiscordMessenger) RespondWithMessage(interaction *discordgo.Interaction, message string) error {
	args := m.Called(interaction, message)
	return args.Error(0)
}

func (m *MockDiscordMessenger) SendPlayStatus(channelID string, playMsg *entity.PlayedSong) (messageID string, err error) {
	args := m.Called(channelID, playMsg)
	return args.String(0), args.Error(1)
}

func (m *MockDiscordMessenger) UpdatePlayStatus(channelID, messageID string, playMsg *entity.PlayedSong) error {
	args := m.Called(channelID, messageID, playMsg)
	return args.Error(0)
}

func (m *MockDiscordMessenger) Respond(interaction *discordgo.Interaction, response *discordgo.InteractionResponse) error {
	args := m.Called(interaction, response)
	return args.Error(0)
}

func (m *MockDiscordMessenger) EditMessageByID(channelID, messageID string, content string) error {
	args := m.Called(channelID, messageID, content)
	return args.Error(0)
}

func (m *MockDiscordMessenger) GetOriginalResponseID(interaction *discordgo.Interaction) (string, error) {
	args := m.Called(interaction)
	if args.Get(0) == nil {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
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
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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

type MockPlayRequestService struct {
	mock.Mock
}

func (m *MockPlayRequestService) Enqueue(guildID string, data model.PlayRequestData) <-chan model.PlayResult {
	args := m.Called(guildID, data)
	return args.Get(0).(<-chan model.PlayResult)
}
