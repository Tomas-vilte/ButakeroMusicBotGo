////go:build !integration

package player

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/voice"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/errors_app"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"sync"
	"testing"
	"time"
)

var errPlaylistEmpty = errors_app.NewAppError(errors_app.ErrCodePlaylistEmpty, "No hay canciones disponibles en la playlist", nil)

func TestAddSongConcurrently(t *testing.T) {
	ctx := context.Background()

	guildPlayer1, mockSongStorage1, mockStateStorage1, mockVoiceSession1, mockPlaybackHandler1, mockLogger1 := setupGuildPlayer("server1")
	guildPlayer2, mockSongStorage2, mockStateStorage2, mockVoiceSession2, mockPlaybackHandler2, mockLogger2 := setupGuildPlayer("server2")

	server1TextChannel := "text-channel-server1"
	server1VoiceChannel := "voice-channel-server1"
	server2TextChannel := "text-channel-server2"
	server2VoiceChannel := "voice-channel-server2"

	song1 := createTestSong("song1", "Song 1 Title")
	song2 := createTestSong("song2", "Song 2 Title")

	mockLogger1.On("With", mock.Anything, mock.Anything).Return(mockLogger1)
	mockLogger1.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger1.On("Debug", mock.Anything, mock.Anything).Return()

	mockLogger2.On("With", mock.Anything, mock.Anything).Return(mockLogger2)
	mockLogger2.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger2.On("Debug", mock.Anything, mock.Anything).Return()

	mockStateStorage1.On("SetTextChannelID", mock.Anything, server1TextChannel).Return(nil)
	mockStateStorage1.On("SetVoiceChannelID", mock.Anything, server1VoiceChannel).Return(nil)
	mockStateStorage1.On("GetVoiceChannelID", mock.Anything).Return(server1VoiceChannel, nil)
	mockStateStorage1.On("GetTextChannelID", mock.Anything).Return(server1TextChannel, nil)
	mockStateStorage1.On("GetCurrentTrack", mock.Anything).Return((*entity.PlayedSong)(nil), nil)

	mockStateStorage2.On("SetTextChannelID", mock.Anything, server2TextChannel).Return(nil)
	mockStateStorage2.On("SetVoiceChannelID", mock.Anything, server2VoiceChannel).Return(nil)
	mockStateStorage2.On("GetVoiceChannelID", mock.Anything).Return(server2VoiceChannel, nil)
	mockStateStorage2.On("GetTextChannelID", mock.Anything).Return(server2TextChannel, nil)
	mockStateStorage2.On("GetCurrentTrack", mock.Anything).Return((*entity.PlayedSong)(nil), nil)

	mockSongStorage1.On("AppendTrack", mock.Anything, song1).Return(nil)
	mockSongStorage2.On("AppendTrack", mock.Anything, song2).Return(nil)

	mockSongStorage1.On("ClearPlaylist", mock.Anything).Return(nil)
	mockSongStorage2.On("ClearPlaylist", mock.Anything).Return(nil)

	mockPlaybackHandler1.On("Stop", mock.Anything).Return(nil)
	mockPlaybackHandler2.On("Stop", mock.Anything).Return(nil)

	mockPlaybackHandler1.On("CurrentState").Return(StateIdle)
	mockPlaybackHandler2.On("CurrentState").Return(StateIdle)

	mockVoiceSession1.On("JoinVoiceChannel", mock.Anything, server1VoiceChannel).Return(nil)
	mockVoiceSession2.On("JoinVoiceChannel", mock.Anything, server2VoiceChannel).Return(nil)

	mockSongStorage1.On("PopNextTrack", mock.Anything).Return((*entity.PlayedSong)(nil), errPlaylistEmpty)
	mockSongStorage2.On("PopNextTrack", mock.Anything).Return((*entity.PlayedSong)(nil), errPlaylistEmpty)

	mockVoiceSession1.On("LeaveVoiceChannel", mock.Anything).Return(nil)
	mockVoiceSession2.On("LeaveVoiceChannel", mock.Anything).Return(nil)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		err := guildPlayer1.AddSong(ctx, &server1TextChannel, &server1VoiceChannel, song1)
		assert.NoError(t, err)
	}()

	go func() {
		defer wg.Done()
		err := guildPlayer2.AddSong(ctx, &server2TextChannel, &server2VoiceChannel, song2)
		assert.NoError(t, err)
	}()

	wg.Wait()

	time.Sleep(100 * time.Millisecond)

	mockSongStorage1.AssertCalled(t, "AppendTrack", mock.Anything, song1)
	mockSongStorage1.AssertNotCalled(t, "AppendTrack", mock.Anything, song2)

	mockSongStorage2.AssertCalled(t, "AppendTrack", mock.Anything, song2)
	mockSongStorage2.AssertNotCalled(t, "AppendTrack", mock.Anything, song1)

	mockVoiceSession1.AssertCalled(t, "JoinVoiceChannel", mock.Anything, server1VoiceChannel)
	mockVoiceSession1.AssertNotCalled(t, "JoinVoiceChannel", mock.Anything, server2VoiceChannel)

	mockVoiceSession2.AssertCalled(t, "JoinVoiceChannel", mock.Anything, server2VoiceChannel)
	mockVoiceSession2.AssertNotCalled(t, "JoinVoiceChannel", mock.Anything, server1VoiceChannel)

	_ = guildPlayer1.Close()
	_ = guildPlayer2.Close()
}

func TestAddSongRaceCondition(t *testing.T) {
	ctx := context.Background()

	guildPlayer1, mockSongStorage1, mockStateStorage1, mockVoiceSession1, mockPlaybackHandler1, mockLogger1 := setupGuildPlayer("server1")

	server1TextChannel := "text-channel-server1"
	server1VoiceChannel := "voice-channel-server1"
	wrongVoiceChannel := "voice-channel-wrong"

	mockLogger1.On("With", mock.Anything, mock.Anything).Return(mockLogger1)
	mockLogger1.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger1.On("Debug", mock.Anything, mock.Anything).Return()
	mockSongStorage1.On("ClearPlaylist", mock.Anything).Return(nil)
	mockPlaybackHandler1.On("Stop", mock.Anything).Return(nil)

	song1 := createTestSong("song1", "Song 1 Title")

	mockStateStorage1.On("SetTextChannelID", mock.Anything, server1TextChannel).Return(nil)
	mockStateStorage1.On("SetVoiceChannelID", mock.Anything, server1VoiceChannel).Return(nil)

	mockStateStorage1.On("GetVoiceChannelID", mock.Anything).Return(wrongVoiceChannel, nil).Once()
	mockStateStorage1.On("GetTextChannelID", mock.Anything).Return(server1TextChannel, nil)
	mockStateStorage1.On("GetCurrentTrack", mock.Anything).Return((*entity.PlayedSong)(nil), nil)

	mockSongStorage1.On("AppendTrack", mock.Anything, song1).Return(nil)

	mockPlaybackHandler1.On("CurrentState").Return(StateIdle)

	mockVoiceSession1.On("JoinVoiceChannel", mock.Anything, wrongVoiceChannel).Return(nil)
	mockSongStorage1.On("PopNextTrack", mock.Anything).Return((*entity.PlayedSong)(nil), errPlaylistEmpty)
	mockVoiceSession1.On("LeaveVoiceChannel", mock.Anything).Return(nil)

	err := guildPlayer1.AddSong(ctx, &server1TextChannel, &server1VoiceChannel, song1)

	time.Sleep(100 * time.Millisecond)

	assert.NoError(t, err)
	mockStateStorage1.AssertCalled(t, "SetVoiceChannelID", mock.Anything, server1VoiceChannel)

	mockVoiceSession1.AssertCalled(t, "JoinVoiceChannel", mock.Anything, wrongVoiceChannel)
	mockVoiceSession1.AssertNotCalled(t, "JoinVoiceChannel", mock.Anything, server1VoiceChannel)

	_ = guildPlayer1.Close()
}

func TestConcurrentAddAndSkip(t *testing.T) {
	ctx := context.Background()

	guildPlayer, mockPlaylistHandler, mockStateStorage, mockVoiceSession, mockPlaybackHandler, mockLogger := setupGuildPlayer("server1")
	guildPlayer.playbackHandler = mockPlaybackHandler

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()

	mockPlaylistHandler.On("ClearPlaylist", mock.Anything).Return(nil)

	textChannel := "text-channel-1"
	voiceChannel := "voice-channel-1"
	song := createTestSong("song-id", "Test Song")

	mockStateStorage.On("SetTextChannelID", mock.Anything, textChannel).Return(nil)
	mockStateStorage.On("SetVoiceChannelID", mock.Anything, voiceChannel).Return(nil)
	mockStateStorage.On("GetVoiceChannelID", mock.Anything).Return(voiceChannel, nil)
	mockStateStorage.On("GetTextChannelID", mock.Anything).Return(textChannel, nil)

	currentSong := &entity.PlayedSong{
		DiscordSong: &entity.DiscordEntity{
			ID:         "song-id",
			TitleTrack: "Test Song",
		},
		Position: 0,
	}

	mockStateStorage.On("GetCurrentTrack", mock.Anything).Return(currentSong, nil)

	mockPlaylistHandler.On("AppendTrack", mock.Anything, song).Return(nil)
	mockPlaylistHandler.On("PopNextTrack", mock.Anything).Return(song, nil).Once()
	mockPlaylistHandler.On("PopNextTrack", mock.Anything).Return(nil, errPlaylistEmpty).After(time.Millisecond * 500)

	mockVoiceSession.On("JoinVoiceChannel", mock.Anything, voiceChannel).Return(nil)
	mockVoiceSession.On("LeaveVoiceChannel", mock.Anything).Return(nil)

	mockPlaybackHandler.On("Play", mock.Anything, song, textChannel).Return(nil)
	mockPlaybackHandler.On("CurrentState").Return(StateIdle)
	mockPlaybackHandler.On("Stop", mock.Anything).Return(nil)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		err := guildPlayer.AddSong(ctx, &textChannel, &voiceChannel, song)
		assert.NoError(t, err)
	}()

	time.Sleep(50 * time.Millisecond)

	go func() {
		defer wg.Done()
		guildPlayer.SkipSong(ctx)
	}()

	wg.Wait()
	time.Sleep(100 * time.Millisecond)

	mockPlaybackHandler.AssertCalled(t, "Stop", mock.Anything)

	_ = guildPlayer.Close()
}

func TestPauseResume(t *testing.T) {
	ctx := context.Background()
	guildPlayer, _, mockStateStorage, _, mockPlaybackHandler, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	currentSong := &entity.PlayedSong{
		DiscordSong: &entity.DiscordEntity{
			ID:         "song-id",
			TitleTrack: "Test Song",
		},
		Position: 50,
	}

	mockStateStorage.On("GetCurrentTrack", mock.Anything).Return(currentSong, nil)
	mockPlaybackHandler.On("Pause", mock.Anything).Return(nil)
	mockPlaybackHandler.On("Resume", mock.Anything).Return(nil)

	pauseErr := guildPlayer.Pause(ctx)
	resumeErr := guildPlayer.Resume(ctx)

	assert.NoError(t, pauseErr)
	assert.NoError(t, resumeErr)
	mockPlaybackHandler.AssertCalled(t, "Pause", mock.Anything)
	mockPlaybackHandler.AssertCalled(t, "Resume", mock.Anything)
}

func TestPauseError(t *testing.T) {
	ctx := context.Background()
	guildPlayer, _, mockStateStorage, _, mockPlaybackHandler, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	currentSong := &entity.PlayedSong{
		DiscordSong: &entity.DiscordEntity{
			ID:         "song-id",
			TitleTrack: "Test Song",
		},
		Position: 50,
	}

	expectedErr := errors.New("error al pausar")
	mockStateStorage.On("GetCurrentTrack", mock.Anything).Return(currentSong, nil)
	mockPlaybackHandler.On("Pause", mock.Anything).Return(expectedErr)

	err := guildPlayer.Pause(ctx)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockPlaybackHandler.AssertCalled(t, "Pause", mock.Anything)
}

func TestResumeError(t *testing.T) {
	ctx := context.Background()
	guildPlayer, _, mockStateStorage, _, mockPlaybackHandler, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	currentSong := &entity.PlayedSong{
		DiscordSong: &entity.DiscordEntity{
			ID:         "song-id",
			TitleTrack: "Test Song",
		},
		Position: 50,
	}

	expectedErr := errors.New("error al reanudar")
	mockStateStorage.On("GetCurrentTrack", mock.Anything).Return(currentSong, nil)
	mockPlaybackHandler.On("Resume", mock.Anything).Return(expectedErr)

	err := guildPlayer.Resume(ctx)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockPlaybackHandler.AssertCalled(t, "Resume", mock.Anything)
}

func TestStop(t *testing.T) {
	ctx := context.Background()
	guildPlayer, mockPlaylistHandler, mockStateStorage, mockVoiceSession, mockPlaybackHandler, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	currentSong := &entity.PlayedSong{
		DiscordSong: &entity.DiscordEntity{
			ID:         "song-id",
			TitleTrack: "Test Song",
		},
		Position: 50,
	}

	mockStateStorage.On("GetCurrentTrack", mock.Anything).Return(currentSong, nil)
	mockPlaylistHandler.On("ClearPlaylist", mock.Anything).Return(nil)
	mockPlaybackHandler.On("Stop", mock.Anything).Return()
	mockVoiceSession.On("LeaveVoiceChannel", mock.Anything).Return(nil)

	err := guildPlayer.Stop(ctx)

	assert.NoError(t, err)
	mockPlaylistHandler.AssertCalled(t, "ClearPlaylist", mock.Anything)
	mockPlaybackHandler.AssertCalled(t, "Stop", mock.Anything)
	mockVoiceSession.AssertCalled(t, "LeaveVoiceChannel", mock.Anything)
}

func TestStopWithPlaylistClearError(t *testing.T) {
	ctx := context.Background()
	guildPlayer, mockPlaylistHandler, mockStateStorage, _, mockPlaybackHandler, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	currentSong := &entity.PlayedSong{
		DiscordSong: &entity.DiscordEntity{
			ID:         "song-id",
			TitleTrack: "Test Song",
		},
		Position: 50,
	}

	expectedErr := errors.New("error al limpiar playlist")
	mockStateStorage.On("GetCurrentTrack", mock.Anything).Return(currentSong, nil)
	mockPlaylistHandler.On("ClearPlaylist", mock.Anything).Return(expectedErr)
	mockPlaybackHandler.On("Stop", mock.Anything).Return()

	err := guildPlayer.Stop(ctx)

	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "error al limpiar la lista"))
	mockPlaylistHandler.AssertCalled(t, "ClearPlaylist", mock.Anything)
	mockPlaybackHandler.AssertNotCalled(t, "Stop", mock.Anything)
}

func TestRemoveSongPlayer(t *testing.T) {
	ctx := context.Background()
	guildPlayer, mockPlaylistHandler, _, _, _, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	position := 2
	removedSong := &entity.PlayedSong{
		DiscordSong: &entity.DiscordEntity{
			ID:         "removed-song",
			TitleTrack: "Removed Song",
		},
	}

	mockPlaylistHandler.On("RemoveTrack", mock.Anything, position).Return(removedSong, nil)

	song, err := guildPlayer.RemoveSong(ctx, position)

	assert.NoError(t, err)
	assert.Equal(t, removedSong, song)
	mockPlaylistHandler.AssertCalled(t, "RemoveTrack", mock.Anything, position)
}

func TestRemoveSongError(t *testing.T) {
	ctx := context.Background()
	guildPlayer, mockPlaylistHandler, _, _, _, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	position := 2
	expectedErr := errors.New("canción no encontrada")

	mockPlaylistHandler.On("RemoveTrack", mock.Anything, position).Return(nil, expectedErr)

	song, err := guildPlayer.RemoveSong(ctx, position)

	assert.Error(t, err)
	assert.Nil(t, song)
	assert.True(t, strings.Contains(err.Error(), "error al remover la cancion"))
	mockPlaylistHandler.AssertCalled(t, "RemoveTrack", mock.Anything, position)
}

func TestGetPlaylistPlayer(t *testing.T) {
	ctx := context.Background()
	guildPlayer, mockPlaylistHandler, _, _, _, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	expectedPlaylist := []*entity.PlayedSong{
		{DiscordSong: &entity.DiscordEntity{
			ID: "Canción 1 - Artista 1",
		}},
		{DiscordSong: &entity.DiscordEntity{
			ID: "Canción 2 - Artista 2",
		}},
		{DiscordSong: &entity.DiscordEntity{
			ID: "Canción 3 - Artista 3",
		}},
	}

	mockPlaylistHandler.On("GetAllTracks", mock.Anything).Return(expectedPlaylist, nil)

	playlist, err := guildPlayer.GetPlaylist(ctx)

	assert.NoError(t, err)
	assert.Equal(t, expectedPlaylist, playlist)
	mockPlaylistHandler.AssertCalled(t, "GetAllTracks", mock.Anything)
}

func TestGetPlaylistError(t *testing.T) {
	ctx := context.Background()
	guildPlayer, mockPlaylistHandler, _, _, _, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	expectedErr := errors.New("error al obtener playlist")
	mockPlaylistHandler.On("GetAllTracks", mock.Anything).Return([]*entity.PlayedSong{}, expectedErr)

	playlist, err := guildPlayer.GetPlaylist(ctx)

	assert.Error(t, err)
	assert.Nil(t, playlist)
	assert.True(t, strings.Contains(err.Error(), "error al obtener la playlist"))
	mockPlaylistHandler.AssertCalled(t, "GetAllTracks", mock.Anything)
}

func TestGetPlayedSong(t *testing.T) {
	ctx := context.Background()
	guildPlayer, _, mockStateStorage, _, _, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	currentSong := &entity.PlayedSong{
		DiscordSong: &entity.DiscordEntity{
			ID:         "song-id",
			TitleTrack: "Test Song",
		},
		Position: 50,
	}

	mockStateStorage.On("GetCurrentTrack", mock.Anything).Return(currentSong, nil)

	song, err := guildPlayer.GetPlayedSong(ctx)

	assert.NoError(t, err)
	assert.Equal(t, currentSong, song)
	mockStateStorage.AssertCalled(t, "GetCurrentTrack", mock.Anything)
}

func TestGetPlayedSongWithoutCurrentSong(t *testing.T) {
	ctx := context.Background()
	guildPlayer, _, mockStateStorage, _, _, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockStateStorage.On("GetCurrentTrack", mock.Anything).Return((*entity.PlayedSong)(nil), nil)

	song, err := guildPlayer.GetPlayedSong(ctx)

	assert.NoError(t, err)
	assert.Nil(t, song)
	mockStateStorage.AssertCalled(t, "GetCurrentTrack", mock.Anything)
}

func TestGetPlayedSongError(t *testing.T) {
	ctx := context.Background()
	guildPlayer, _, mockStateStorage, _, _, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	expectedErr := errors.New("error al obtener canción actual")
	mockStateStorage.On("GetCurrentTrack", mock.Anything).Return(&entity.PlayedSong{}, expectedErr)

	song, err := guildPlayer.GetPlayedSong(ctx)

	assert.Error(t, err)
	assert.Nil(t, song)
	assert.True(t, strings.Contains(err.Error(), "error al obtener la cancion actual"))
	mockStateStorage.AssertCalled(t, "GetCurrentTrack", mock.Anything)
}

func TestJoinVoiceChannel(t *testing.T) {
	ctx := context.Background()
	guildPlayer, _, _, mockVoiceSession, _, _ := setupGuildPlayer("server1")

	channelID := "voice-channel-id"
	mockVoiceSession.On("JoinVoiceChannel", mock.Anything, channelID).Return(nil)

	err := guildPlayer.JoinVoiceChannel(ctx, channelID)

	assert.NoError(t, err)
	mockVoiceSession.AssertCalled(t, "JoinVoiceChannel", mock.Anything, channelID)
}

func TestJoinVoiceChannelError(t *testing.T) {
	ctx := context.Background()
	guildPlayer, _, _, mockVoiceSession, _, _ := setupGuildPlayer("server1")

	channelID := "voice-channel-id"
	expectedErr := voice.ErrNoVoiceConnection
	mockVoiceSession.On("JoinVoiceChannel", mock.Anything, channelID).Return(expectedErr)

	err := guildPlayer.JoinVoiceChannel(ctx, channelID)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockVoiceSession.AssertCalled(t, "JoinVoiceChannel", mock.Anything, channelID)
}

func TestRunWithContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	guildPlayer, _, mockStateStorage, _, _, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

	mockStateStorage.On("GetCurrentTrack", mock.Anything).Return((*entity.PlayedSong)(nil), nil)

	go func() {
		cancel()
	}()
	err := guildPlayer.Run(ctx)

	assert.Error(t, err)
	assert.Equal(t, ctx.Err(), err)
}

func TestRestoreCurrentTrack(t *testing.T) {
	ctx := context.Background()
	guildPlayer, mockPlaylistHandler, mockStateStorage, _, _, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	currentSong := &entity.PlayedSong{
		DiscordSong: &entity.DiscordEntity{
			ID:         "song-id",
			TitleTrack: "Test Song",
		},
		Position:      50,
		StartPosition: 0,
	}

	expectedRestoredSong := &entity.PlayedSong{
		DiscordSong: &entity.DiscordEntity{
			ID:         "song-id",
			TitleTrack: "Test Song",
		},
		Position:      50,
		StartPosition: 50,
	}

	mockStateStorage.On("GetCurrentTrack", mock.Anything).Return(currentSong, nil)
	mockPlaylistHandler.On("AppendTrack", mock.Anything, mock.MatchedBy(func(s *entity.PlayedSong) bool {
		return s.StartPosition == 50 && s.DiscordSong.ID == "song-id"
	})).Return(nil)

	err := guildPlayer.restoreCurrentTrack(ctx)

	assert.NoError(t, err)
	mockStateStorage.AssertCalled(t, "GetCurrentTrack", mock.Anything)
	mockPlaylistHandler.AssertCalled(t, "AppendTrack", mock.Anything, mock.MatchedBy(func(s *entity.PlayedSong) bool {
		return s.StartPosition == expectedRestoredSong.StartPosition &&
			s.DiscordSong.ID == expectedRestoredSong.DiscordSong.ID
	}))
}

func TestHandleEvent(t *testing.T) {
	ctx := context.Background()
	guildPlayer, _, _, _, _, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything, mock.Anything).Return()

	mockEvent := new(MockPlayerEvent)
	mockEvent.On("Type").Return(EventType("test_event"))
	mockEvent.On("HandleEvent", mock.Anything, mock.Anything).Return(nil)

	// Act
	err := guildPlayer.handleEvent(ctx, mockEvent)

	assert.NoError(t, err)
	mockEvent.AssertCalled(t, "HandleEvent", mock.Anything, mock.Anything)
}

func TestClose(t *testing.T) {
	guildPlayer, mockPlaylistHandler, mockStateStorage, mockVoiceSession, mockPlaybackHandler, mockLogger := setupGuildPlayer("server1")

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockStateStorage.On("GetCurrentTrack", mock.Anything).Return((*entity.PlayedSong)(nil), nil)
	mockPlaylistHandler.On("ClearPlaylist", mock.Anything).Return(nil)
	mockPlaybackHandler.On("Stop", mock.Anything).Return()
	mockVoiceSession.On("LeaveVoiceChannel", mock.Anything).Return(nil)

	err := guildPlayer.Close()

	assert.NoError(t, err)
	mockPlaylistHandler.AssertCalled(t, "ClearPlaylist", mock.Anything)
	mockPlaybackHandler.AssertCalled(t, "Stop", mock.Anything)
	mockVoiceSession.AssertCalled(t, "LeaveVoiceChannel", mock.Anything)
}

func setupGuildPlayer(_ string) (*GuildPlayer, *MockSongStorage, *MockPlayerStateStorage, *MockVoiceSession, *MockPlaybackHandler, *logging.MockLogger) {
	mockSongStorage := new(MockSongStorage)
	mockPlaybackHandler := new(MockPlaybackHandler)
	mockVoiceSession := new(MockVoiceSession)
	mockStateStorage := new(MockPlayerStateStorage)
	logger := new(logging.MockLogger)

	guildPlayer := &GuildPlayer{
		playbackHandler: mockPlaybackHandler,
		voiceSession:    mockVoiceSession,
		stateStorage:    mockStateStorage,
		songStorage:     mockSongStorage,
		eventCh:         make(chan PlayerEvent, 100),
		logger:          logger,
	}

	return guildPlayer, mockSongStorage, mockStateStorage, mockVoiceSession, mockPlaybackHandler, logger
}

func createTestSong(id, title string) *entity.PlayedSong {
	return &entity.PlayedSong{
		DiscordSong: &entity.DiscordEntity{
			ID:         id,
			TitleTrack: title,
		},
		Position: 0,
	}
}
