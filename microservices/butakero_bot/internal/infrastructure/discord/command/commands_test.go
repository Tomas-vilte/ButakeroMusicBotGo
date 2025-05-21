//go:build !integration

package command

import (
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/errors_app"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestCommandHandler_PlaySong_Success(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)

	session := &discordgo.Session{State: discordgo.NewState()}
	session.State.User = &discordgo.User{ID: "bot123"}

	guild := &discordgo.Guild{
		ID: "guild123",
		VoiceStates: []*discordgo.VoiceState{
			{UserID: "user123", ChannelID: "voiceChannel123"},
		},
	}
	err := session.State.GuildAdd(guild)
	if err != nil {
		t.Fatalf("Error al añadir el estado del servidor: %v", err)
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	opt := &discordgo.ApplicationCommandInteractionDataOption{
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{
				Type:  discordgo.ApplicationCommandOptionString,
				Name:  "song",
				Value: "test song",
			},
		},
	}

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockDiscordMessenger.On("Respond", mock.Anything, mock.Anything).Return(nil)
	mockDiscordMessenger.On("GetOriginalResponseID", mock.Anything).Return("msg123", nil)
	mockDiscordMessenger.On("EditMessageByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	resultChan := make(chan model.PlayResult, 1)
	readOnlyChan := (<-chan model.PlayResult)(resultChan)

	mockQueueManager.On("Enqueue", mock.Anything, mock.Anything).Return(readOnlyChan)

	go func() {
		resultChan <- model.PlayResult{
			SongTitle: "Test Song Title",
			Err:       nil,
		}
		close(resultChan)
	}()

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	// Act
	handler.PlaySong(session, interaction, opt)

	time.Sleep(100 * time.Millisecond)

	// Assert
	mockLogger.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
	mockQueueManager.AssertExpectations(t)

	mockQueueManager.AssertCalled(t, "Enqueue", "guild123", mock.MatchedBy(func(data model.PlayRequestData) bool {
		return data.GuildID == "guild123" &&
			data.SongInput == "test song" &&
			data.VoiceChannelID == "voiceChannel123" &&
			data.RequestedByName == "testUser"
	}))
}

func TestCommandHandler_StopPlaying_Success(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)
	mockGuildPlayer := new(MockGuildPlayer)

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockGuildManager.On("GetGuildPlayer", "guild123").Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("Stop", mock.Anything).Return(nil)
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, SuccessMessagePlayingStopped).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	// Act
	handler.StopPlaying(interaction)

	// Assert
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
}

func TestCommandHandler_ListPlaylist_ErrorGettingPlaylist(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)
	mockGuildPlayer := new(MockGuildPlayer)

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Error", "Error al obtener la lista de reproducción", mock.Anything).Return()
	mockGuildManager.On("GetGuildPlayer", "guild123").Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("GetPlaylist", mock.Anything).Return([]*entity.PlayedSong{}, errors.New("error getting playlist"))
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, ErrorMessageGenericPlaylist).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	// Act
	handler.ListPlaylist(interaction)

	// Assert
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestCommandHandler_RemoveSong_InvalidPosition(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Warn", "Opción de posición para remover canción inválida o faltante", mock.Anything).Return()
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, ErrorMessageInvalidRemovePosition).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	opt := &discordgo.ApplicationCommandInteractionDataOption{
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{
				Type:  discordgo.ApplicationCommandOptionString, // Wrong type
				Name:  "posición",
				Value: "invalid",
			},
		},
	}

	// Act
	handler.RemoveSong(interaction, opt)

	// Assert
	mockLogger.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
}

func TestCommandHandler_GetPlayingSong_ErrorGettingSong(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)
	mockGuildPlayer := new(MockGuildPlayer)

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Error", "Error al obtener la canción actual", mock.Anything).Return()
	mockGuildManager.On("GetGuildPlayer", "guild123").Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("GetPlayedSong", mock.Anything).Return(nil, errors.New("error getting song"))
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, ErrorMessageGenericPlaylist).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	// Act
	handler.GetPlayingSong(interaction)

	// Assert
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestCommandHandler_PauseSong_Error(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)
	mockGuildPlayer := new(MockGuildPlayer)

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockGuildManager.On("GetGuildPlayer", "guild123").Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("Pause", mock.Anything).Return(errors.New("failed to pause"))
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, ErrorMessageGenericPause).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	// Act
	handler.PauseSong(interaction)

	// Assert
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestCommandHandler_PlaySong_Error_QueueEnqueue(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)

	session := &discordgo.Session{State: discordgo.NewState()}
	session.State.User = &discordgo.User{ID: "bot123"}

	guild := &discordgo.Guild{
		ID: "guild123",
		VoiceStates: []*discordgo.VoiceState{
			{UserID: "user123", ChannelID: "voiceChannel123"},
		},
	}
	err := session.State.GuildAdd(guild)
	if err != nil {
		t.Fatalf("Error al añadir el estado del servidor: %v", err)
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	opt := &discordgo.ApplicationCommandInteractionDataOption{
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{
				Type:  discordgo.ApplicationCommandOptionString,
				Name:  "song",
				Value: "test song",
			},
		},
	}

	resultChan := make(chan model.PlayResult, 1)
	readOnlyChan := (<-chan model.PlayResult)(resultChan)

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockDiscordMessenger.On("Respond", mock.Anything, mock.Anything).Return(nil)
	mockDiscordMessenger.On("GetOriginalResponseID", mock.Anything).Return("msg123", nil)
	mockDiscordMessenger.On("EditMessageByID", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockQueueManager.On("Enqueue", mock.Anything, mock.Anything).Return(readOnlyChan)

	go func() {
		resultChan <- model.PlayResult{
			SongTitle: "",
			Err:       errors.New("error al encolar canción"),
		}
		close(resultChan)
	}()

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	// Act
	handler.PlaySong(session, interaction, opt)

	time.Sleep(100 * time.Millisecond)

	// Assert
	mockQueueManager.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestCommandHandler_PlaySong_Error_InitialResponse(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)

	session := &discordgo.Session{State: discordgo.NewState()}
	session.State.User = &discordgo.User{ID: "bot123"}

	guild := &discordgo.Guild{
		ID: "guild123",
		VoiceStates: []*discordgo.VoiceState{
			{UserID: "user123", ChannelID: "voiceChannel123"},
		},
	}
	err := session.State.GuildAdd(guild)
	if err != nil {
		t.Fatalf("Error al añadir el estado del servidor: %v", err)
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	opt := &discordgo.ApplicationCommandInteractionDataOption{
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{
				Type:  discordgo.ApplicationCommandOptionString,
				Name:  "song",
				Value: "test song",
			},
		},
	}

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", "Error al enviar respuesta inicial", mock.Anything).Return()
	mockDiscordMessenger.On("Respond", mock.Anything, mock.Anything).Return(errors.New("error al responder"))

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	// Act
	handler.PlaySong(session, interaction, opt)

	// Assert
	mockDiscordMessenger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestCommandHandler_StopPlaying_FailedToGetPlayer(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)
	mockGuildPlayer := new(MockGuildPlayer)

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Error", "Error al obtener GuildPlayer", mock.Anything).Return()
	mockGuildManager.On("GetGuildPlayer", "guild123").Return(mockGuildPlayer, errors.New("player not found"))
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, ErrorMessageGuildPlayerNotAccesible).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	// Act
	handler.StopPlaying(interaction)

	// Assert
	mockGuildManager.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestCommandHandler_StopPlaying_FailedToStop(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)
	mockGuildPlayer := new(MockGuildPlayer)

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Error", "Error al detener la reproducción", mock.Anything).Return()
	mockGuildManager.On("GetGuildPlayer", "guild123").Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("Stop", mock.Anything).Return(errors.New("failed to stop"))
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, ErrorMessageGenericStop).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	// Act
	handler.StopPlaying(interaction)

	// Assert
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestCommandHandler_SkipSong_Success(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)
	mockGuildPlayer := new(MockGuildPlayer)

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockGuildManager.On("GetGuildPlayer", "guild123").Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("SkipSong", mock.Anything).Return(nil)
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, SuccessMessageSongSkipped).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	// Act
	handler.SkipSong(interaction)

	// Assert
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
}

func TestCommandHandler_SkipSong_NotPlaying(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)
	mockGuildPlayer := new(MockGuildPlayer)

	appErr := &errors_app.AppError{
		Code:    errors_app.ErrCodePlayerNotPlaying,
		Message: "No hay canción reproduciéndose",
	}

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", "Intento de skip pero no había canción reproduciéndose.", mock.Anything).Return()
	mockGuildManager.On("GetGuildPlayer", "guild123").Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("SkipSong", mock.Anything).Return(appErr)
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, ErrorMessageNothingToSkip).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	// Act
	handler.SkipSong(interaction)

	// Assert
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestCommandHandler_ListPlaylist_Success(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)
	mockGuildPlayer := new(MockGuildPlayer)

	songs := []*entity.PlayedSong{
		{DiscordSong: &entity.DiscordEntity{TitleTrack: "Canción 1"}},
		{DiscordSong: &entity.DiscordEntity{TitleTrack: "Canción 2"}},
	}

	expectedEmbed := &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{
				{Title: "🎵 Lista de reproducción:", Description: "1. Canción 1\n2. Canción 2\n"},
			},
		},
	}

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", "Mostrando lista de reproducción", mock.Anything).Return()
	mockGuildManager.On("GetGuildPlayer", "guild123").Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("GetPlaylist", mock.Anything).Return(songs, nil)
	mockDiscordMessenger.On("Respond", mock.Anything, mock.MatchedBy(func(resp *discordgo.InteractionResponse) bool {
		return resp.Type == expectedEmbed.Type &&
			len(resp.Data.Embeds) == 1 &&
			resp.Data.Embeds[0].Title == expectedEmbed.Data.Embeds[0].Title
	})).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	// Act
	handler.ListPlaylist(interaction)

	// Assert
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
}

func TestCommandHandler_ListPlaylist_Empty(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)
	mockGuildPlayer := new(MockGuildPlayer)

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockGuildManager.On("GetGuildPlayer", "guild123").Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("GetPlaylist", mock.Anything).Return([]*entity.PlayedSong{}, nil)
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, InfoMessagePlaylistEmpty).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	// Act
	handler.ListPlaylist(interaction)

	// Assert
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
}

func TestCommandHandler_RemoveSong_Success(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)
	mockGuildPlayer := new(MockGuildPlayer)

	song := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "Canción Removida"}}

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", "Canción eliminada exitosamente", mock.Anything).Return()
	mockGuildManager.On("GetGuildPlayer", "guild123").Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("RemoveSong", mock.Anything, 2).Return(song, nil)
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, fmt.Sprintf(SuccessMessageSongRemovedFmt, song.DiscordSong.TitleTrack)).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	opt := &discordgo.ApplicationCommandInteractionDataOption{
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{
				Type:  discordgo.ApplicationCommandOptionInteger,
				Name:  "posición",
				Value: float64(2),
			},
		},
	}

	// Act
	handler.RemoveSong(interaction, opt)

	// Assert
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestCommandHandler_GetPlayingSong_Success(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)
	mockGuildPlayer := new(MockGuildPlayer)

	song := &entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "Canción Actual"}}

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", "Mostrando canción actual", mock.Anything).Return()
	mockGuildManager.On("GetGuildPlayer", "guild123").Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("GetPlayedSong", mock.Anything).Return(song, nil)
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, fmt.Sprintf(InfoMessageNowPlayingFmt, song.DiscordSong.TitleTrack)).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	// Act
	handler.GetPlayingSong(interaction)

	// Assert
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestCommandHandler_GetPlayingSong_NoSongPlaying(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)
	mockGuildPlayer := new(MockGuildPlayer)

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockGuildManager.On("GetGuildPlayer", "guild123").Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("GetPlayedSong", mock.Anything).Return(nil, nil)
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, ErrorMessageNoCurrentSong).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	// Act
	handler.GetPlayingSong(interaction)

	// Assert
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestCommandHandler_PlaySong_Error_NotInVoiceChannel(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)

	session := &discordgo.Session{State: discordgo.NewState()}
	session.State.User = &discordgo.User{ID: "bot123"}

	// Guild sin estado de voz para el usuario
	guild := &discordgo.Guild{
		ID:          "guild123",
		VoiceStates: []*discordgo.VoiceState{},
	}
	err := session.State.GuildAdd(guild)
	if err != nil {
		t.Fatalf("Error al añadir el estado del servidor: %v", err)
	}

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	opt := &discordgo.ApplicationCommandInteractionDataOption{
		Options: []*discordgo.ApplicationCommandInteractionDataOption{
			{
				Type:  discordgo.ApplicationCommandOptionString,
				Name:  "song",
				Value: "test song",
			},
		},
	}

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, ErrorMessageNotInVoiceChannel).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	// Act
	handler.PlaySong(session, interaction, opt)

	// Assert
	mockDiscordMessenger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestCommandHandler_PauseSong_Success(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)
	mockGuildPlayer := new(MockGuildPlayer)

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockGuildManager.On("GetGuildPlayer", "guild123").Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("Pause", mock.Anything).Return(nil)
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, SuccessMessagePaused).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	// Act
	handler.PauseSong(interaction)

	// Assert
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestCommandHandler_ResumeSong_Success(t *testing.T) {
	// Arrange
	mockStorage := new(MockInteractionStorage)
	mockLogger := new(logging.MockLogger)
	mockGuildManager := new(MockGuildManager)
	mockDiscordMessenger := new(MockDiscordMessenger)
	mockQueueManager := new(MockPlayRequestService)
	mockGuildPlayer := new(MockGuildPlayer)

	mockLogger.On("With", mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockGuildManager.On("GetGuildPlayer", "guild123").Return(mockGuildPlayer, nil)
	mockGuildPlayer.On("Resume", mock.Anything).Return(nil)
	mockDiscordMessenger.On("RespondWithMessage", mock.Anything, SuccessMessageResumed).Return(nil)

	handler := NewCommandHandler(mockStorage, mockLogger, mockGuildManager, mockDiscordMessenger, mockQueueManager)

	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			GuildID:   "guild123",
			ChannelID: "channel123",
			Member: &discordgo.Member{
				User: &discordgo.User{
					ID:       "user123",
					Username: "testUser",
				},
			},
		},
	}

	// Act
	handler.ResumeSong(interaction)

	// Assert
	mockGuildManager.AssertExpectations(t)
	mockGuildPlayer.AssertExpectations(t)
	mockDiscordMessenger.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
