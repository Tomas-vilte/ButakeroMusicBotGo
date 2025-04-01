package discord

import (
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// VoiceStateService maneja la lógica de estado de voz
type VoiceStateService struct {
	logger logging.Logger
}

func NewVoiceStateService(logger logging.Logger) *VoiceStateService {
	return &VoiceStateService{
		logger: logger,
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

	var botVoiceState *discordgo.VoiceState
	for _, state := range guild.VoiceStates {
		if state.UserID == session.State.User.ID {
			botVoiceState = state
			break
		}
	}

	if botVoiceState != nil {
		wasInBotChannel := vs.BeforeUpdate != nil && vs.BeforeUpdate.ChannelID == botVoiceState.ChannelID

		movedToDifferentChannel := vs.ChannelID != "" && vs.ChannelID != botVoiceState.ChannelID

		if wasInBotChannel && movedToDifferentChannel {
			newChannelID := vs.ChannelID

			usersInCurrentChannel := 0
			for _, state := range guild.VoiceStates {
				if state.UserID != session.State.User.ID && state.ChannelID == botVoiceState.ChannelID {
					usersInCurrentChannel++
				}
			}

			if usersInCurrentChannel == 0 {
				stateStorage := guildPlayer.StateStorage()

				if err := stateStorage.SetVoiceChannel(newChannelID); err != nil {
					return fmt.Errorf("error al actualizar el canal de voz: %w", err)
				}

				if err := guildPlayer.JoinVoiceChannel(newChannelID); err != nil {
					return fmt.Errorf("error al mover el bot al nuevo canal: %w", err)
				}

				s.logger.Info("Bot movido al nuevo canal",
					zap.String("oldChannel", botVoiceState.ChannelID),
					zap.String("newChannel", newChannelID),
					zap.String("userID", vs.UserID))

				return nil
			}
		}
	}

	currentSong, err := guildPlayer.GetPlayedSong()
	if err != nil {
		return fmt.Errorf("error al obtener la canción actual: %w", err)
	}

	if currentSong != nil && vs.UserID == currentSong.RequestedByID {
		if vs.ChannelID == "" || (botVoiceState != nil && vs.ChannelID != botVoiceState.ChannelID) {
			usersInCurrentChannel := 0
			for _, state := range guild.VoiceStates {
				if state.UserID != session.State.User.ID &&
					(botVoiceState != nil && state.ChannelID == botVoiceState.ChannelID) {
					usersInCurrentChannel++
				}
			}

			if usersInCurrentChannel == 0 {
				if err := guildPlayer.Stop(); err != nil {
					return fmt.Errorf("error al detener la reproducción: %w", err)
				}
			}
		}
	}

	return nil
}
