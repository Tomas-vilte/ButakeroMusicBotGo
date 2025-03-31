package events

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord"
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
	guildsPlayers     map[GuildID]ports.GuildPlayer
	cfg               *config.Config
	logger            logging.Logger
	discordMessenger  ports.DiscordMessenger
	storageAudio      ports.StorageAudio
	voiceStateService *discord.VoiceStateService
}

// NewEventHandler crea una nueva instancia de EventHandler.
func NewEventHandler(
	cfg *config.Config,
	logger logging.Logger,
	discordMessenger ports.DiscordMessenger,
	storageAudio ports.StorageAudio,
	voiceStateService *discord.VoiceStateService,
) *EventHandler {
	return &EventHandler{
		guildsPlayers:     make(map[GuildID]ports.GuildPlayer),
		cfg:               cfg,
		logger:            logger,
		discordMessenger:  discordMessenger,
		storageAudio:      storageAudio,
		voiceStateService: voiceStateService,
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
	if vs.UserID == s.State.User.ID {
		return
	}

	guildID := GuildID(vs.GuildID)
	guildPlayer, exists := h.guildsPlayers[guildID]
	if !exists {
		return
	}

	if err := h.voiceStateService.HandleVoiceStateChange(guildPlayer, s, vs); err != nil {
		h.logger.Error("Error al manejar cambio de estado de voz", zap.Error(err))
	}
}
