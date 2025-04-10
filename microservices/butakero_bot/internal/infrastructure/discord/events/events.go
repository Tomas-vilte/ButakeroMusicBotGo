package events

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// GuildID representa el ID de un servidor de Discord.
type GuildID string

// EventHandler maneja los eventos de Discord.
type EventHandler struct {
	guildManager      ports.GuildManager
	cfg               *config.Config
	logger            logging.Logger
	voiceStateService discord.VoiceStateHandler
}

func NewEventHandler(
	guildManager ports.GuildManager,
	voiceStateService discord.VoiceStateHandler,
	logger logging.Logger,
	cfg *config.Config,
) *EventHandler {
	return &EventHandler{
		guildManager:      guildManager,
		voiceStateService: voiceStateService,
		logger:            logger,
		cfg:               cfg,
	}
}

// Ready se llama cuando el bot está listo para recibir interacciones.
func (h *EventHandler) Ready(s *discordgo.Session, _ *discordgo.Ready) {
	if err := s.UpdateGameStatus(0, fmt.Sprintf("con tu vieja /%s", h.cfg.CommandPrefix)); err != nil {
		h.logger.Error("Error al actualizar el estado del juego", zap.Error(err))
	}
}

// GuildCreate se llama cuando el bot se une a un nuevo servidor.
func (h *EventHandler) GuildCreate(ctx context.Context, _ *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	if _, err := h.guildManager.GetGuildPlayer(event.Guild.ID); err == nil {
		h.logger.Debug("GuildPlayer ya existe", zap.String("guildID", event.Guild.ID))
		return
	}

	guildPlayer, err := h.guildManager.CreateGuildPlayer(event.Guild.ID)
	if err != nil {
		h.logger.Error("Error al crear GuildPlayer", zap.String("guildID", event.Guild.ID), zap.Error(err))
		return
	}

	go func() {
		if err := guildPlayer.Run(ctx); err != nil {
			h.logger.Error("Error al iniciar GuildPlayer", zap.String("guildID", event.Guild.ID), zap.Error(err))
		}
	}()
}

// GuildDelete se llama cuando el bot es removido de un servidor.
func (h *EventHandler) GuildDelete(_ *discordgo.Session, event *discordgo.GuildDelete) {
	if err := h.guildManager.RemoveGuildPlayer(event.Guild.ID); err != nil {
		h.logger.Error("Error al eliminar GuildPlayer", zap.String("guildID", event.Guild.ID), zap.Error(err))
	}
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

	guildPlayer, err := h.guildManager.GetGuildPlayer(vs.GuildID)
	if err != nil {
		h.logger.Debug("GuildPlayer no encontrado", zap.String("guildID", vs.GuildID))
		return
	}

	if err := h.voiceStateService.HandleVoiceStateChange(guildPlayer, s, vs); err != nil {
		h.logger.Error("Error al manejar cambio de voz", zap.Error(err))
	}
}
