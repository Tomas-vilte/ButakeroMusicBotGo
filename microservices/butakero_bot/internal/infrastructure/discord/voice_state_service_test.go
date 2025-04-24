//go:build !integration

package discord

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestBotChannelTracker_GetBotVoiceState(t *testing.T) {
	logger := new(logging.MockLogger)
	tracker := NewBotChannelTracker(logger)

	// Escenario: Obtener el estado de voz del bot cuando está en un canal
	guild := &discordgo.Guild{
		VoiceStates: []*discordgo.VoiceState{
			{UserID: "bot123", ChannelID: "channel1"},
			{UserID: "user1", ChannelID: "channel2"},
		},
	}
	session := &discordgo.Session{State: discordgo.NewState()}
	session.State.User = &discordgo.User{ID: "bot123"}

	result := tracker.GetBotVoiceState(guild, session)

	assert.NotNil(t, result)
	assert.Equal(t, "bot123", result.UserID)
	assert.Equal(t, "channel1", result.ChannelID)
}

func TestBotChannelTracker_CountUsersInChannel(t *testing.T) {
	logger := new(logging.MockLogger)
	tracker := NewBotChannelTracker(logger)

	// Escenario: Contar usuarios en un canal excluyendo al usuario especificado
	guild := &discordgo.Guild{
		VoiceStates: []*discordgo.VoiceState{
			{UserID: "user1", ChannelID: "channel1"},
			{UserID: "user2", ChannelID: "channel1"},
			{UserID: "user3", ChannelID: "channel2"},
		},
	}

	count := tracker.CountUsersInChannel(guild, "channel1", "user1")

	assert.Equal(t, 1, count)
}

func TestBotMover_MoveBotToNewChannel_Success(t *testing.T) {
	logger := new(logging.MockLogger)
	mover := NewBotMover(logger)

	// Escenario: Mover el bot a un nuevo canal de voz exitosamente
	guildPlayer := new(MockGuildPlayer)
	stateStorage := new(MockStateStorage)

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	guildPlayer.On("StateStorage").Return(stateStorage)
	stateStorage.On("SetVoiceChannel", "newChannel").Return(nil)
	guildPlayer.On("JoinVoiceChannel", "newChannel").Return(nil)

	err := mover.MoveBotToNewChannel(guildPlayer, "newChannel", "oldChannel", "user123")

	assert.NoError(t, err)
	guildPlayer.AssertExpectations(t)
	stateStorage.AssertExpectations(t)
}

func TestPlaybackController_HandlePlayback_StopWhenRequesterLeaves(t *testing.T) {
	logger := new(logging.MockLogger)
	controller := NewPlaybackController(logger)

	// Escenario: Detener la reproducción cuando el usuario que solicitó la canción abandona el canal
	guildPlayer := new(MockGuildPlayer)
	session := &discordgo.Session{State: discordgo.NewState()}
	session.State.User = &discordgo.User{ID: "bot123"}

	currentSong := &entity.PlayedSong{RequestedByID: "user123", DiscordSong: &entity.DiscordEntity{TitleTrack: "titleTrack"}}
	vs := &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			UserID:    "user123",
			ChannelID: "",
			GuildID:   "guild123",
		},
		BeforeUpdate: &discordgo.VoiceState{
			ChannelID: "channel123",
		},
	}
	botVoiceState := &discordgo.VoiceState{ChannelID: "channel123", UserID: "bot123"}
	guild := &discordgo.Guild{
		VoiceStates: []*discordgo.VoiceState{
			botVoiceState,
		},
	}

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	guildPlayer.On("Stop").Return(nil)

	err := controller.HandlePlayback(guildPlayer, currentSong, vs, botVoiceState, guild, session)

	assert.NoError(t, err)
	guildPlayer.AssertExpectations(t)
}

func TestVoiceStateService_HandleVoiceStateChange_BotNotInVoiceChannel(t *testing.T) {
	logger := new(logging.MockLogger)
	tracker := NewBotChannelTracker(logger)
	mover := NewBotMover(logger)
	playback := NewPlaybackController(logger)
	service := NewVoiceStateService(tracker, mover, playback, logger)

	// Escenario: Cambio de estado de voz cuando el bot no está en ningún canal
	guildPlayer := new(MockGuildPlayer)
	session := &discordgo.Session{State: discordgo.NewState()}
	session.State.User = &discordgo.User{ID: "bot123"}

	guild := &discordgo.Guild{
		ID:          "guild123",
		VoiceStates: []*discordgo.VoiceState{},
	}
	err := session.State.GuildAdd(guild)
	if err != nil {
		t.Fatalf("Error al añadir el estado del servidor: %v", err)
		return
	}

	vs := &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			UserID:    "user123",
			ChannelID: "channel2",
			GuildID:   "guild123",
		},
		BeforeUpdate: &discordgo.VoiceState{
			ChannelID: "channel1",
		},
	}

	currentSong := &entity.PlayedSong{RequestedByID: "user123", DiscordSong: &entity.DiscordEntity{TitleTrack: "titleTrack"}}
	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	guildPlayer.On("GetPlayedSong").Return(currentSong, nil)

	err = service.HandleVoiceStateChange(guildPlayer, session, vs)

	assert.NoError(t, err)
	guildPlayer.AssertExpectations(t)
}

func TestHandleVoiceStateChange_BotMovesWhenChannelEmpty(t *testing.T) {
	logger := new(logging.MockLogger)
	tracker := NewBotChannelTracker(logger)
	mover := NewBotMover(logger)
	playback := NewPlaybackController(logger)
	service := NewVoiceStateService(tracker, mover, playback, logger)

	// Escenario: El bot se mueve cuando el usuario abandona dejando el canal vacío
	guildPlayer := new(MockGuildPlayer)
	stateStorage := new(MockStateStorage)
	session := &discordgo.Session{State: discordgo.NewState()}
	session.State.User = &discordgo.User{ID: "bot123"}

	guild := &discordgo.Guild{
		ID: "guild123",
		VoiceStates: []*discordgo.VoiceState{
			{UserID: "bot123", ChannelID: "channel1"},
		},
	}
	err := session.State.GuildAdd(guild)
	if err != nil {
		t.Fatalf("Error al añadir el estado del servidor: %v", err)
		return
	}

	vs := &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			UserID:    "user123",
			ChannelID: "channel2",
			GuildID:   "guild123",
		},
		BeforeUpdate: &discordgo.VoiceState{
			ChannelID: "channel1",
		},
	}

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	guildPlayer.On("StateStorage").Return(stateStorage)
	stateStorage.On("SetVoiceChannel", "channel2").Return(nil)
	guildPlayer.On("JoinVoiceChannel", "channel2").Return(nil)

	err = service.HandleVoiceStateChange(guildPlayer, session, vs)

	assert.NoError(t, err)
	guildPlayer.AssertExpectations(t)
	stateStorage.AssertExpectations(t)
}

func TestHandleVoiceStateChange_StopsPlaybackWhenRequesterLeaves(t *testing.T) {
	logger := new(logging.MockLogger)
	tracker := NewBotChannelTracker(logger)
	mover := NewBotMover(logger)
	playback := NewPlaybackController(logger)
	service := NewVoiceStateService(tracker, mover, playback, logger)

	// Escenario: Detener la reproducción cuando el solicitante abandona el canal
	guildPlayer := new(MockGuildPlayer)
	session := &discordgo.Session{State: discordgo.NewState()}
	session.State.User = &discordgo.User{ID: "bot123"}

	guild := &discordgo.Guild{
		ID: "guild123",
		VoiceStates: []*discordgo.VoiceState{
			{UserID: "bot123", ChannelID: "channel1"},
		},
	}
	err := session.State.GuildAdd(guild)
	if err != nil {
		t.Fatalf("Error al añadir el estado del servidor: %v", err)
		return
	}

	vs := &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			UserID:    "user123",
			ChannelID: "",
			GuildID:   "guild123",
		},
		BeforeUpdate: &discordgo.VoiceState{
			ChannelID: "channel1",
		},
	}

	currentSong := &entity.PlayedSong{RequestedByID: "user123", DiscordSong: &entity.DiscordEntity{TitleTrack: "titleTrack"}}
	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Info", mock.Anything, mock.Anything).Return()
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	guildPlayer.On("GetPlayedSong").Return(currentSong, nil)
	guildPlayer.On("Stop").Return(nil)

	err = service.HandleVoiceStateChange(guildPlayer, session, vs)

	assert.NoError(t, err)
	guildPlayer.AssertExpectations(t)
}

func TestHandleVoiceStateChange_BotDoesNotMoveWhenChannelNotEmpty(t *testing.T) {
	logger := new(logging.MockLogger)
	tracker := NewBotChannelTracker(logger)
	mover := NewBotMover(logger)
	playback := NewPlaybackController(logger)
	service := NewVoiceStateService(tracker, mover, playback, logger)

	// Escenario: El bot no se mueve cuando quedan otros usuarios en el canal
	guildPlayer := new(MockGuildPlayer)
	session := &discordgo.Session{State: discordgo.NewState()}
	session.State.User = &discordgo.User{ID: "bot123"}

	guild := &discordgo.Guild{
		ID: "guild123",
		VoiceStates: []*discordgo.VoiceState{
			{UserID: "bot123", ChannelID: "channel1"},
			{UserID: "user456", ChannelID: "channel1"},
		},
	}
	err := session.State.GuildAdd(guild)
	if err != nil {
		t.Fatalf("Error al añadir el estado del servidor: %v", err)
		return
	}

	vs := &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			UserID:    "user123",
			ChannelID: "channel2",
			GuildID:   "guild123",
		},
		BeforeUpdate: &discordgo.VoiceState{
			ChannelID: "channel1",
		},
	}

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	guildPlayer.On("GetPlayedSong").Return(&entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "titleTrack"}}, nil)

	err = service.HandleVoiceStateChange(guildPlayer, session, vs)

	assert.NoError(t, err)
	guildPlayer.AssertNotCalled(t, "JoinVoiceChannel")
}

func TestHandleVoiceStateChange_ErrorGettingGuild(t *testing.T) {
	logger := new(logging.MockLogger)
	tracker := NewBotChannelTracker(logger)
	mover := NewBotMover(logger)
	playback := NewPlaybackController(logger)
	service := NewVoiceStateService(tracker, mover, playback, logger)

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Error", mock.Anything, mock.Anything).Return()

	// Escenario: Error al obtener el servidor del estado de la sesión
	guildPlayer := new(MockGuildPlayer)
	session := &discordgo.Session{State: discordgo.NewState()}

	vs := &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			UserID:    "user123",
			ChannelID: "channel2",
			GuildID:   "guild123",
		},
	}

	err := service.HandleVoiceStateChange(guildPlayer, session, vs)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al obtener el servidor")
}

func TestHandleVoiceStateChange_UserWasNotInBotChannel(t *testing.T) {
	logger := new(logging.MockLogger)
	tracker := NewBotChannelTracker(logger)
	mover := NewBotMover(logger)
	playback := NewPlaybackController(logger)
	service := NewVoiceStateService(tracker, mover, playback, logger)

	// Escenario: Usuario cambia de canal pero nunca estuvo en el canal del bot
	guildPlayer := new(MockGuildPlayer)
	session := &discordgo.Session{State: discordgo.NewState()}
	session.State.User = &discordgo.User{ID: "bot123"}

	guild := &discordgo.Guild{
		ID: "guild123",
		VoiceStates: []*discordgo.VoiceState{
			{UserID: "bot123", ChannelID: "channel1"},
		},
	}
	err := session.State.GuildAdd(guild)
	if err != nil {
		t.Fatalf("Error al añadir el estado del servidor: %v", err)
		return
	}

	vs := &discordgo.VoiceStateUpdate{
		VoiceState: &discordgo.VoiceState{
			UserID:    "user123",
			ChannelID: "channel2",
			GuildID:   "guild123",
		},
		BeforeUpdate: &discordgo.VoiceState{
			ChannelID: "channel3",
		},
	}

	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	guildPlayer.On("GetPlayedSong").Return(&entity.PlayedSong{DiscordSong: &entity.DiscordEntity{TitleTrack: "titleTrack"}}, nil)

	err = service.HandleVoiceStateChange(guildPlayer, session, vs)

	assert.NoError(t, err)
	guildPlayer.AssertNotCalled(t, "JoinVoiceChannel")
	guildPlayer.AssertNotCalled(t, "Stop")
}
