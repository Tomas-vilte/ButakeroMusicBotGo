package player

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/mock"
	"io"
)

type MockPlaylistStorage struct {
	mock.Mock
}

func (m *MockPlaylistStorage) AppendTrack(ctx context.Context, song *entity.PlayedSong) error {
	args := m.Called(ctx, song)
	return args.Error(0)
}

func (m *MockPlaylistStorage) RemoveTrack(ctx context.Context, position int) (*entity.PlayedSong, error) {
	args := m.Called(ctx, position)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PlayedSong), args.Error(1)
}

func (m *MockPlaylistStorage) GetAllTracks(ctx context.Context) ([]*entity.PlayedSong, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.PlayedSong), args.Error(1)
}

func (m *MockPlaylistStorage) ClearPlaylist(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPlaylistStorage) PopNextTrack(ctx context.Context) (*entity.PlayedSong, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PlayedSong), args.Error(1)
}

type MockVoiceSession struct {
	mock.Mock
}

func (m *MockVoiceSession) JoinVoiceChannel(ctx context.Context, channelID string) error {
	args := m.Called(ctx, channelID)
	return args.Error(0)
}

func (m *MockVoiceSession) LeaveVoiceChannel(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockVoiceSession) SendAudio(ctx context.Context, reader io.ReadCloser) error {
	args := m.Called(ctx, reader)
	return args.Error(0)
}

func (m *MockVoiceSession) Pause() {
	m.Called()
}

func (m *MockVoiceSession) Resume() {
	m.Called()
}

type MockStorageAudio struct {
	mock.Mock
}

func (m *MockStorageAudio) GetAudio(ctx context.Context, songID string) (io.ReadCloser, error) {
	args := m.Called(ctx, songID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.ReadCloser), args.Error(1)
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

type MockPlayerStateStorage struct {
	mock.Mock
}

func (m *MockPlayerStateStorage) GetCurrentTrack(ctx context.Context) (*entity.PlayedSong, error) {
	args := m.Called(ctx)
	return args.Get(0).(*entity.PlayedSong), args.Error(1)
}

func (m *MockPlayerStateStorage) SetCurrentTrack(ctx context.Context, track *entity.PlayedSong) error {
	args := m.Called(ctx, track)
	return args.Error(0)
}

func (m *MockPlayerStateStorage) GetVoiceChannelID(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockPlayerStateStorage) SetVoiceChannelID(ctx context.Context, channelID string) error {
	args := m.Called(ctx, channelID)
	return args.Error(0)
}

func (m *MockPlayerStateStorage) GetTextChannelID(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockPlayerStateStorage) SetTextChannelID(ctx context.Context, channelID string) error {
	args := m.Called(ctx, channelID)
	return args.Error(0)

}

type MockPlaylistHandler struct {
	mock.Mock
}

func (m *MockPlaylistHandler) AddSong(ctx context.Context, song *entity.PlayedSong) error {
	args := m.Called(ctx, song)
	return args.Error(0)
}

func (m *MockPlaylistHandler) RemoveSong(ctx context.Context, position int) (*entity.DiscordEntity, error) {
	args := m.Called(ctx, position)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.DiscordEntity), args.Error(1)
}

func (m *MockPlaylistHandler) GetPlaylist(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockPlaylistHandler) GetNextSong(ctx context.Context) (*entity.PlayedSong, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.PlayedSong), args.Error(1)
}

func (m *MockPlaylistHandler) ClearPlaylist(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

type MockPlaybackHandler struct {
	mock.Mock
}

func (m *MockPlaybackHandler) Play(ctx context.Context, song *entity.PlayedSong, textChannelID string) error {
	args := m.Called(ctx, song, textChannelID)
	return args.Error(0)
}

func (m *MockPlaybackHandler) Stop(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockPlaybackHandler) Pause(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPlaybackHandler) Resume(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPlaybackHandler) CurrentState() PlayerState {
	args := m.Called()
	return args.Get(0).(PlayerState)
}

type MockPlayerEvent struct {
	mock.Mock
}

func (m *MockPlayerEvent) Type() EventType {
	args := m.Called()
	return args.Get(0).(EventType)
}

func (m *MockPlayerEvent) HandleEvent(ctx context.Context, player *GuildPlayer) error {
	args := m.Called(ctx, player)
	return args.Error(0)
}
