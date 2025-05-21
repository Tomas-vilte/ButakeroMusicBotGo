//go:build !integration

package service

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestEnqueue_Success(t *testing.T) {
	// arrange
	mockSongService := new(MockSongService)
	mockGuildManager := new(MockGuildManager)
	mockLogger := new(logging.MockLogger)
	mockGuildPlayer := new(MockGuildPlayer)

	prm := NewPlayRequestManager(mockSongService, mockGuildManager, mockLogger)

	ctx := context.Background()
	guildID := "123456789"
	userID := "987654321"
	songInput := "test song"
	channelID := "channel123"
	voiceChannelID := "voice123"
	requestedByName := "Test User"

	discordSong := &entity.DiscordEntity{
		TitleTrack: "Test Song Title",
		DurationMs: 3600000,
		URL:        "https://example.com/test",
	}

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockSongService.On("GetOrDownloadSong", mock.Anything, userID, songInput, "youtube").Return(discordSong, nil)
	mockGuildManager.On("GetGuildPlayer", guildID).Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("AddSong", mock.Anything, &channelID, &voiceChannelID, mock.MatchedBy(func(s *entity.PlayedSong) bool {
		return s.RequestedByID == userID && s.RequestedByName == requestedByName && s.DiscordSong.TitleTrack == discordSong.TitleTrack
	})).Return(nil)

	requestData := model.PlayRequestData{
		Ctx:             ctx,
		GuildID:         guildID,
		UserID:          userID,
		ChannelID:       channelID,
		VoiceChannelID:  voiceChannelID,
		SongInput:       songInput,
		RequestedByName: requestedByName,
	}

	// act
	resultChan := prm.Enqueue(guildID, requestData)

	// assert
	result := <-resultChan
	assert.NoError(t, result.Err)
	assert.Equal(t, discordSong.TitleTrack, result.SongTitle)
	assert.Equal(t, userID, result.RequestedByID)
	assert.Equal(t, requestedByName, result.RequestedByName)

	mockSongService.AssertExpectations(t)
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
}

func TestEnqueue_SongServiceError(t *testing.T) {
	// arrange
	mockSongService := new(MockSongService)
	mockGuildManager := new(MockGuildManager)
	mockLogger := new(logging.MockLogger)

	prm := NewPlayRequestManager(mockSongService, mockGuildManager, mockLogger)

	ctx := context.Background()
	guildID := "123456789"
	userID := "987654321"
	songInput := "test song"
	channelID := "channel123"
	voiceChannelID := "voice123"
	requestedByName := "Test User"
	expectedError := errors.New("song download failed")

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockSongService.On("GetOrDownloadSong", mock.Anything, userID, songInput, "youtube").Return(nil, expectedError)

	requestData := model.PlayRequestData{
		Ctx:             ctx,
		GuildID:         guildID,
		UserID:          userID,
		ChannelID:       channelID,
		VoiceChannelID:  voiceChannelID,
		SongInput:       songInput,
		RequestedByName: requestedByName,
	}

	// act
	resultChan := prm.Enqueue(guildID, requestData)

	// assert
	result := <-resultChan
	assert.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "no se pudo obtener/descargar la canción")
	assert.Equal(t, userID, result.RequestedByID)
	assert.Equal(t, requestedByName, result.RequestedByName)

	mockSongService.AssertExpectations(t)
}

func TestEnqueue_GuildManagerError(t *testing.T) {
	// arrange
	mockSongService := new(MockSongService)
	mockGuildManager := new(MockGuildManager)
	mockLogger := new(logging.MockLogger)
	mockGuildPlayer := new(MockGuildPlayer)

	prm := NewPlayRequestManager(mockSongService, mockGuildManager, mockLogger)

	ctx := context.Background()
	guildID := "123456789"
	userID := "987654321"
	songInput := "test song"
	channelID := "channel123"
	voiceChannelID := "voice123"
	requestedByName := "Test User"
	expectedError := errors.New("guild player not found")

	discordSong := &entity.DiscordEntity{
		TitleTrack: "Test Song Title",
		DurationMs: 3600000,
		URL:        "https://example.com/test",
	}

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockSongService.On("GetOrDownloadSong", mock.Anything, userID, songInput, "youtube").Return(discordSong, nil)
	mockGuildManager.On("GetGuildPlayer", guildID).Return(mockGuildPlayer, expectedError)

	requestData := model.PlayRequestData{
		Ctx:             ctx,
		GuildID:         guildID,
		UserID:          userID,
		ChannelID:       channelID,
		VoiceChannelID:  voiceChannelID,
		SongInput:       songInput,
		RequestedByName: requestedByName,
	}

	// act
	resultChan := prm.Enqueue(guildID, requestData)

	// assert
	result := <-resultChan
	assert.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "error al obtener GuildPlayer")
	assert.Equal(t, userID, result.RequestedByID)
	assert.Equal(t, requestedByName, result.RequestedByName)

	mockSongService.AssertExpectations(t)
	mockGuildManager.AssertExpectations(t)
}

func TestEnqueue_AddSongError(t *testing.T) {
	// arrange
	mockSongService := new(MockSongService)
	mockGuildManager := new(MockGuildManager)
	mockLogger := new(logging.MockLogger)
	mockGuildPlayer := new(MockGuildPlayer)

	prm := NewPlayRequestManager(mockSongService, mockGuildManager, mockLogger)

	ctx := context.Background()
	guildID := "123456789"
	userID := "987654321"
	songInput := "test song"
	channelID := "channel123"
	voiceChannelID := "voice123"
	requestedByName := "Test User"
	expectedError := errors.New("could not add song to queue")

	discordSong := &entity.DiscordEntity{
		TitleTrack: "Test Song Title",
		DurationMs: 36000000,
		URL:        "https://example.com/test",
	}

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockSongService.On("GetOrDownloadSong", mock.Anything, userID, songInput, "youtube").Return(discordSong, nil)
	mockGuildManager.On("GetGuildPlayer", guildID).Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("AddSong", mock.Anything, &channelID, &voiceChannelID, mock.MatchedBy(func(s *entity.PlayedSong) bool {
		return s.RequestedByID == userID && s.RequestedByName == requestedByName && s.DiscordSong.TitleTrack == discordSong.TitleTrack
	})).Return(expectedError)

	requestData := model.PlayRequestData{
		Ctx:             ctx,
		GuildID:         guildID,
		UserID:          userID,
		ChannelID:       channelID,
		VoiceChannelID:  voiceChannelID,
		SongInput:       songInput,
		RequestedByName: requestedByName,
	}

	// act
	resultChan := prm.Enqueue(guildID, requestData)

	// assert
	result := <-resultChan
	assert.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "no se pudo agregar la canción")
	assert.Equal(t, discordSong.TitleTrack, result.SongTitle)
	assert.Equal(t, userID, result.RequestedByID)
	assert.Equal(t, requestedByName, result.RequestedByName)

	mockSongService.AssertExpectations(t)
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
}

func TestNewPlayRequestManager(t *testing.T) {
	// arrange
	mockSongService := new(MockSongService)
	mockGuildManager := new(MockGuildManager)
	mockLogger := new(logging.MockLogger)

	// act
	prm := NewPlayRequestManager(mockSongService, mockGuildManager, mockLogger)

	// assert
	assert.NotNil(t, prm)
	assert.NotNil(t, prm.guildQueues)
	assert.Equal(t, mockSongService, prm.songService)
	assert.Equal(t, mockGuildManager, prm.guildManager)
	assert.Equal(t, mockLogger, prm.logger)
}

func TestMultipleEnqueueRequests(t *testing.T) {
	// arrange
	mockSongService := new(MockSongService)
	mockGuildManager := new(MockGuildManager)
	mockLogger := new(logging.MockLogger)
	mockGuildPlayer := new(MockGuildPlayer)

	prm := NewPlayRequestManager(mockSongService, mockGuildManager, mockLogger)

	ctx := context.Background()
	guildID := "123456789"
	userID := "987654321"
	channelID := "channel123"
	voiceChannelID := "voice123"
	requestedByName := "Test User"

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	song1Input := "first song"
	song1 := &entity.DiscordEntity{
		TitleTrack: "First Song Title",
		DurationMs: 3600000,
		URL:        "https://example.com/first",
	}

	mockSongService.On("GetOrDownloadSong", mock.Anything, userID, song1Input, "youtube").Return(song1, nil)
	mockGuildManager.On("GetGuildPlayer", guildID).Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("AddSong", mock.Anything, &channelID, &voiceChannelID, mock.MatchedBy(func(s *entity.PlayedSong) bool {
		return s.DiscordSong.TitleTrack == song1.TitleTrack
	})).Return(nil)

	song2Input := "second song"
	song2 := &entity.DiscordEntity{
		TitleTrack: "Second Song Title",
		DurationMs: 4800000,
		URL:        "https://example.com/second",
	}

	mockSongService.On("GetOrDownloadSong", mock.Anything, userID, song2Input, "youtube").Return(song2, nil)
	mockGuildPlayer.On("AddSong", mock.Anything, &channelID, &voiceChannelID, mock.MatchedBy(func(s *entity.PlayedSong) bool {
		return s.DiscordSong.TitleTrack == song2.TitleTrack
	})).Return(nil)

	request1 := model.PlayRequestData{
		Ctx:             ctx,
		GuildID:         guildID,
		UserID:          userID,
		ChannelID:       channelID,
		VoiceChannelID:  voiceChannelID,
		SongInput:       song1Input,
		RequestedByName: requestedByName,
	}

	request2 := model.PlayRequestData{
		Ctx:             ctx,
		GuildID:         guildID,
		UserID:          userID,
		ChannelID:       channelID,
		VoiceChannelID:  voiceChannelID,
		SongInput:       song2Input,
		RequestedByName: requestedByName,
	}

	// act
	resultChan1 := prm.Enqueue(guildID, request1)
	resultChan2 := prm.Enqueue(guildID, request2)

	// assert
	var result1, result2 model.PlayResult

	select {
	case result1 = <-resultChan1:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for first result")
	}

	select {
	case result2 = <-resultChan2:
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for second result")
	}

	assert.NoError(t, result1.Err)
	assert.Equal(t, song1.TitleTrack, result1.SongTitle)
	assert.Equal(t, userID, result1.RequestedByID)
	assert.Equal(t, requestedByName, result1.RequestedByName)

	assert.NoError(t, result2.Err)
	assert.Equal(t, song2.TitleTrack, result2.SongTitle)
	assert.Equal(t, userID, result2.RequestedByID)
	assert.Equal(t, requestedByName, result2.RequestedByName)

	mockSongService.AssertExpectations(t)
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
}
