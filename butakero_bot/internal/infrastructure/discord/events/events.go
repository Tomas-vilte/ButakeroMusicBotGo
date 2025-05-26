package events

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
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
func (h *EventHandler) Ready(s *discordgo.Session, r *discordgo.Ready) {
	logger := h.logger.With(
		zap.String("component", "EventHandler"),
		zap.String("method", "Ready"),
	)

	if err := s.UpdateGameStatus(0, fmt.Sprintf("con %s /%s", r.User.Username, h.cfg.CommandPrefix)); err != nil {
		logger.Error("Error al actualizar el estado del juego",
			zap.Error(err),
			zap.String("status", h.cfg.CommandPrefix))
		return
	}

	logger.Info("Bot listo y estado actualizado")
}

// GuildCreate se llama cuando el bot se une a un nuevo servidor.
func (h *EventHandler) GuildCreate(ctx context.Context, _ *discordgo.Session, event *discordgo.GuildCreate) {
	logger := h.logger.With(
		zap.String("component", "EventHandler"),
		zap.String("method", "GuildCreate"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("guild_id", event.ID),
	)

	if event.Unavailable {
		logger.Debug("Servidor no disponible - ignorando evento")
		return
	}

	_, err := h.guildManager.GetGuildPlayer(event.ID)
	if err != nil {
		logger.Error("Error al obtener GuildPlayer",
			zap.Error(err))
		return
	}
}

// GuildDelete se llama cuando el bot es removido de un servidor.
func (h *EventHandler) GuildDelete(_ *discordgo.Session, event *discordgo.GuildDelete) {
	logger := h.logger.With(
		zap.String("component", "EventHandler"),
		zap.String("method", "GuildDelete"),
		zap.String("guild_id", event.ID),
	)

	if err := h.guildManager.RemoveGuildPlayer(event.ID); err != nil {
		logger.Error("Error al eliminar GuildPlayer",
			zap.Error(err))
		return
	}

	logger.Debug("GuildPlayer eliminado exitosamente")
}

// RegisterEventHandlers registra los manejadores de eventos en la sesión de Discord.
func (h *EventHandler) RegisterEventHandlers(s *discordgo.Session) {
	s.AddHandler(h.Ready)
	s.AddHandler(func(session *discordgo.Session, event *discordgo.GuildCreate) {
		eventCtx := trace.WithTraceID(context.Background())
		h.GuildCreate(eventCtx, session, event)
	})
	s.AddHandler(h.GuildDelete)
	s.AddHandler(func(session *discordgo.Session, event *discordgo.VoiceStateUpdate) {
		eventCtx := trace.WithTraceID(context.Background())
		h.VoiceStateUpdate(eventCtx, session, event)
	})
}

func (h *EventHandler) VoiceStateUpdate(ctx context.Context, s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	logger := h.logger.With(
		zap.String("component", "EventHandler"),
		zap.String("method", "VoiceStateUpdate"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("guild_id", vs.GuildID),
		zap.String("user_id", vs.UserID),
		zap.String("channel_id", vs.ChannelID),
	)

	if vs.UserID == s.State.User.ID {
		logger.Debug("Cambio de estado de voz del bot - ignorando")
		return
	}

	guildPlayer, err := h.guildManager.GetGuildPlayer(vs.GuildID)
	if err != nil {
		logger.Warn("GuildPlayer no encontrado para el servidor")
		return
	}

	if err := h.voiceStateService.HandleVoiceStateChange(ctx, guildPlayer, s, vs); err != nil {
		logger.Error("Error al manejar cambio de estado de voz",
			zap.Error(err))
		return
	}

	logger.Debug("Cambio de estado de voz manejado exitosamente")
}
