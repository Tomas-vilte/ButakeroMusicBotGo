package discord

import (
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/player"
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

	// Buscar el estado de voz del bot
	var botVoiceState *discordgo.VoiceState
	for _, state := range guild.VoiceStates {
		if state.UserID == session.State.User.ID {
			botVoiceState = state
			break
		}
	}

	// Si el bot está en un canal de voz
	if botVoiceState != nil {
		// Verificar si el usuario estaba en el canal del bot antes del cambio
		wasInBotChannel := vs.BeforeUpdate != nil && vs.BeforeUpdate.ChannelID == botVoiceState.ChannelID

		// Verificar si el usuario se movió a un canal diferente
		movedToDifferentChannel := vs.ChannelID != "" && vs.ChannelID != botVoiceState.ChannelID

		// Si el usuario estaba en el canal del bot y ahora está en otro canal
		if wasInBotChannel && movedToDifferentChannel {
			newChannelID := vs.ChannelID

			// Verificar si hay otros usuarios en el canal actual del bot
			usersInCurrentChannel := 0
			for _, state := range guild.VoiceStates {
				if state.UserID != session.State.User.ID && state.ChannelID == botVoiceState.ChannelID {
					usersInCurrentChannel++
				}
			}

			// Mover el bot al nuevo canal si no hay otros usuarios
			if usersInCurrentChannel == 0 {
				// Actualizar el canal de voz en el stateStorage
				if err := guildPlayer.(*player.GuildPlayer).StateStorage().SetVoiceChannel(newChannelID); err != nil {
					return fmt.Errorf("error al actualizar el canal de voz: %w", err)
				}

				// Mover el bot al nuevo canal
				if err := guildPlayer.(*player.GuildPlayer).Session().JoinVoiceChannel(newChannelID); err != nil {
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

	// Verificar si el usuario que se fue es el solicitante de la canción actual
	currentSong, err := guildPlayer.GetPlayedSong()
	if err != nil {
		return fmt.Errorf("error al obtener la canción actual: %w", err)
	}

	if currentSong != nil && vs.UserID == currentSong.RequestedByID {
		// Verificar si el usuario se desconectó o cambió de canal
		if vs.ChannelID == "" || (botVoiceState != nil && vs.ChannelID != botVoiceState.ChannelID) {
			// Verificar si el bot está solo
			usersInCurrentChannel := 0
			for _, state := range guild.VoiceStates {
				if state.UserID != session.State.User.ID &&
					(botVoiceState != nil && state.ChannelID == botVoiceState.ChannelID) {
					usersInCurrentChannel++
				}
			}

			// Si no hay otros usuarios, detener la reproducción
			if usersInCurrentChannel == 0 {
				if err := guildPlayer.Stop(); err != nil {
					return fmt.Errorf("error al detener la reproducción: %w", err)
				}
			}
		}
	}

	return nil
}
