package discord

import (
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type VoiceStateHandler interface {
	HandleVoiceStateChange(guildPlayer ports.GuildPlayer, session *discordgo.Session, vs *discordgo.VoiceStateUpdate) error
}

type BotChannelTracker struct {
	logger logging.Logger
}

func NewBotChannelTracker(logger logging.Logger) *BotChannelTracker {
	return &BotChannelTracker{logger: logger}
}

func (t *BotChannelTracker) GetBotVoiceState(guild *discordgo.Guild, session *discordgo.Session) *discordgo.VoiceState {
	for _, state := range guild.VoiceStates {
		if state.UserID == session.State.User.ID {
			return state
		}
	}
	return nil
}

func (t *BotChannelTracker) CountUsersInChannel(guild *discordgo.Guild, channelID string, excludeUserID string) int {
	count := 0
	for _, state := range guild.VoiceStates {
		if state.UserID != excludeUserID && state.ChannelID == channelID {
			count++
		}
	}
	return count
}

type BotMover struct {
	logger logging.Logger
}

func NewBotMover(logger logging.Logger) *BotMover {
	return &BotMover{logger: logger}
}

func (m *BotMover) MoveBotToNewChannel(guildPlayer ports.GuildPlayer, newChannelID string, oldChannelID string, userID string) error {
	stateStorage := guildPlayer.StateStorage()
	if err := stateStorage.SetVoiceChannel(newChannelID); err != nil {
		return fmt.Errorf("error al actualizar el canal de voz: %w", err)
	}

	if err := guildPlayer.JoinVoiceChannel(newChannelID); err != nil {
		return fmt.Errorf("error al mover el bot al nuevo canal: %w", err)
	}

	m.logger.Info("Bot movido al nuevo canal",
		zap.String("oldChannel", oldChannelID),
		zap.String("newChannel", newChannelID),
		zap.String("userID", userID))

	return nil
}

type PlaybackController struct {
	logger logging.Logger
}

func NewPlaybackController(logger logging.Logger) *PlaybackController {
	return &PlaybackController{logger: logger}
}

func (c *PlaybackController) HandlePlayback(guildPlayer ports.GuildPlayer, currentSong *entity.PlayedSong, vs *discordgo.VoiceStateUpdate, botVoiceState *discordgo.VoiceState, guild *discordgo.Guild, session *discordgo.Session) error {
	if currentSong != nil && vs.UserID == currentSong.RequestedByID {
		if vs.ChannelID == "" || (botVoiceState != nil && vs.ChannelID != botVoiceState.ChannelID) {
			if botVoiceState != nil {
				usersInBotChannel := 0
				for _, state := range guild.VoiceStates {
					if state.UserID != session.State.User.ID && state.ChannelID == botVoiceState.ChannelID {
						usersInBotChannel++
					}
				}

				if usersInBotChannel == 0 {
					c.logger.Info("No hay usuarios en el canal, deteniendo reproducción",
						zap.String("channelID", botVoiceState.ChannelID))
					if err := guildPlayer.Stop(); err != nil {
						return fmt.Errorf("error al detener la reproducción: %w", err)
					}
				}
			}
		}
	}
	return nil
}

type VoiceStateService struct {
	tracker  *BotChannelTracker
	mover    *BotMover
	playback *PlaybackController
}

func NewVoiceStateService(tracker *BotChannelTracker, mover *BotMover, playback *PlaybackController) *VoiceStateService {
	return &VoiceStateService{
		tracker:  tracker,
		mover:    mover,
		playback: playback,
	}
}

func (s *VoiceStateService) HandleVoiceStateChange(
	guildPlayer ports.GuildPlayer,
	session *discordgo.Session,
	vs *discordgo.VoiceStateUpdate,
) error {
	guildID := vs.GuildID
	guild, err := session.State.Guild(guildID)
	if err != nil {
		return fmt.Errorf("error al obtener el servidor: %w", err)
	}

	botVoiceState := s.tracker.GetBotVoiceState(guild, session)

	if botVoiceState != nil {
		wasInBotChannel := vs.BeforeUpdate != nil && vs.BeforeUpdate.ChannelID == botVoiceState.ChannelID
		movedToDifferentChannel := vs.ChannelID != "" && vs.ChannelID != botVoiceState.ChannelID

		if wasInBotChannel && movedToDifferentChannel {
			newChannelID := vs.ChannelID
			usersInCurrentChannel := s.tracker.CountUsersInChannel(guild, botVoiceState.ChannelID, session.State.User.ID)

			if usersInCurrentChannel == 0 {
				return s.mover.MoveBotToNewChannel(guildPlayer, newChannelID, botVoiceState.ChannelID, vs.UserID)
			}
		}
	}

	currentSong, err := guildPlayer.GetPlayedSong()
	if err != nil {
		return fmt.Errorf("error al obtener la canción actual: %w", err)
	}

	return s.playback.HandlePlayback(guildPlayer, currentSong, vs, botVoiceState, guild, session)
}
