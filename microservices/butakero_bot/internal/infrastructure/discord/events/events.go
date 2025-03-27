package events

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/player"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/voice"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/inmemory"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// GuildID representa el ID de un servidor de Discord.
type GuildID string

// EventHandler maneja los eventos de Discord.
type EventHandler struct {
	guildsPlayers    map[GuildID]ports.GuildPlayer
	cfg              *config.Config
	logger           logging.Logger
	discordMessenger ports.DiscordMessenger
	storageAudio     ports.StorageAudio
}

// NewEventHandler crea una nueva instancia de EventHandler.
func NewEventHandler(
	cfg *config.Config,
	logger logging.Logger,
	discordMessenger ports.DiscordMessenger,
	storageAudio ports.StorageAudio,
) *EventHandler {
	return &EventHandler{
		guildsPlayers:    make(map[GuildID]ports.GuildPlayer),
		cfg:              cfg,
		logger:           logger,
		discordMessenger: discordMessenger,
		storageAudio:     storageAudio,
	}
}

// Ready se llama cuando el bot está listo para recibir interacciones.
func (h *EventHandler) Ready(s *discordgo.Session, _ *discordgo.Ready) {
	if err := s.UpdateGameStatus(0, fmt.Sprintf("con tu vieja /%s", h.cfg.CommandPrefix)); err != nil {
		h.logger.Error("Error al actualizar el estado del juego", zap.Error(err))
	}
}

// GuildCreate se llama cuando el bot se une a un nuevo servidor.
func (h *EventHandler) GuildCreate(ctx context.Context, s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}
	guildPlayer := h.setupGuildPlayer(GuildID(event.Guild.ID), s)
	h.guildsPlayers[GuildID(event.Guild.ID)] = guildPlayer
	h.logger.Debug("Conectando al servidor", zap.String("guildID", event.Guild.ID))
	go func() {
		if err := guildPlayer.Run(ctx); err != nil {
			h.logger.Error("Error al ejecutar el reproductor", zap.Error(err))
		}
	}()
}

// GuildDelete se llama cuando el bot es removido de un servidor.
func (h *EventHandler) GuildDelete(_ *discordgo.Session, event *discordgo.GuildDelete) {
	guildID := GuildID(event.Guild.ID)
	delete(h.guildsPlayers, guildID)
}

// setupGuildPlayer configura un reproductor para un servidor dado.
func (h *EventHandler) setupGuildPlayer(guildID GuildID, dg *discordgo.Session) ports.GuildPlayer {
	voiceChat := voice.NewDiscordVoiceSession(dg, string(guildID), h.logger)
	songStorage := inmemory.NewInmemorySongStorage(h.logger)
	stateStorage := inmemory.NewInmemoryStateStorage(h.logger)

	return player.NewGuildPlayer(
		voiceChat,
		songStorage,
		stateStorage,
		h.discordMessenger,
		h.storageAudio,
		h.logger,
	)
}

// GetGuildPlayer obtiene un reproductor para un servidor dado.
func (h *EventHandler) GetGuildPlayer(guildID GuildID, dg *discordgo.Session) ports.GuildPlayer {
	guildPlayer, ok := h.guildsPlayers[guildID]
	if !ok {
		guildPlayer = h.setupGuildPlayer(guildID, dg)
		h.guildsPlayers[guildID] = guildPlayer
	}
	return guildPlayer
}

// RegisterEventHandlers registra los manejadores de eventos en la sesión de Discord.
func (h *EventHandler) RegisterEventHandlers(s *discordgo.Session, ctx context.Context) {
	s.AddHandler(h.Ready)
	s.AddHandler(func(session *discordgo.Session, event *discordgo.GuildCreate) {
		h.GuildCreate(ctx, session, event)
	})
	s.AddHandler(h.GuildDelete)
	s.AddHandler(h.VoiceStateUpdate)
}

func (h *EventHandler) VoiceStateUpdate(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	// Ignorar eventos del propio bot
	if vs.UserID == s.State.User.ID {
		return
	}

	guildID := GuildID(vs.GuildID)
	guildPlayer, exists := h.guildsPlayers[guildID]
	if !exists {
		return
	}

	// Obtener el estado actual del bot en el servidor
	guild, err := s.State.Guild(string(guildID))
	if err != nil {
		h.logger.Error("Error al obtener el servidor", zap.Error(err))
		return
	}

	// Buscar el estado de voz del bot
	var botVoiceState *discordgo.VoiceState
	for _, state := range guild.VoiceStates {
		if state.UserID == s.State.User.ID {
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
				if state.UserID != s.State.User.ID && state.ChannelID == botVoiceState.ChannelID {
					usersInCurrentChannel++
				}
			}

			// Mover el bot al nuevo canal si no hay otros usuarios
			if usersInCurrentChannel == 0 {
				// Actualizar el canal de voz en el stateStorage
				if err := guildPlayer.(*player.GuildPlayer).StateStorage().SetVoiceChannel(newChannelID); err != nil {
					h.logger.Error("Error al actualizar el canal de voz", zap.Error(err))
					return
				}

				// Mover el bot al nuevo canal
				if err := guildPlayer.(*player.GuildPlayer).Session().JoinVoiceChannel(newChannelID); err != nil {
					h.logger.Error("Error al mover el bot al nuevo canal", zap.Error(err))
					return
				}

				h.logger.Info("Bot movido al nuevo canal",
					zap.String("oldChannel", botVoiceState.ChannelID),
					zap.String("newChannel", newChannelID),
					zap.String("userID", vs.UserID))

				return
			}
		}
	}

	// Verificar si el usuario que se fue es el solicitante de la canción actual
	currentSong, err := guildPlayer.GetPlayedSong()
	if err != nil {
		h.logger.Error("Error al obtener la canción actual", zap.Error(err))
		return
	}

	if currentSong != nil && vs.UserID == currentSong.RequestedByID {
		// Verificar si el usuario se desconectó o cambió de canal
		if vs.ChannelID == "" || (botVoiceState != nil && vs.ChannelID != botVoiceState.ChannelID) {
			h.checkIfBotIsAlone(s, guildID)
		}
	}
}

func (h *EventHandler) checkIfBotIsAlone(s *discordgo.Session, guildID GuildID) {
	guild, err := s.State.Guild(string(guildID))
	if err != nil {
		h.logger.Error("Error al obtener el servidor", zap.Error(err))
		return
	}

	var botVoiceState *discordgo.VoiceState
	for _, vs := range guild.VoiceStates {
		if vs.UserID == s.State.User.ID {
			botVoiceState = vs
			break
		}
	}

	if botVoiceState == nil {
		return
	}

	// Obtener la canción actual para verificar el solicitante
	guildPlayer := h.GetGuildPlayer(guildID, s)
	currentSong, err := guildPlayer.GetPlayedSong()
	if err != nil {
		h.logger.Error("Error al obtener la canción actual", zap.Error(err))
		return
	}

	// Verificar usuarios en el canal
	var requesterInChannel bool
	otherUsersInChannel := 0

	for _, vs := range guild.VoiceStates {
		if vs.UserID != s.State.User.ID && vs.ChannelID == botVoiceState.ChannelID {
			otherUsersInChannel++
			// Verificar si el solicitante sigue en el canal
			if currentSong != nil && vs.UserID == currentSong.RequestedByID {
				requesterInChannel = true
			}
		}
	}

	// Si el solicitante no está en el canal y no hay otros usuarios
	if (currentSong == nil || !requesterInChannel) && otherUsersInChannel == 0 {
		if err := guildPlayer.Stop(); err != nil {
			h.logger.Error("Error al detener la reproducción", zap.Error(err))
		}
	}
}

func (h *EventHandler) Messenger() ports.DiscordMessenger {
	return h.discordMessenger
}
