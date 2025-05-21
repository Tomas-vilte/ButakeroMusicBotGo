package discord

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type VoiceStateHandler interface {
	HandleVoiceStateChange(ctx context.Context, guildPlayer ports.GuildPlayer, session *discordgo.Session, vs *discordgo.VoiceStateUpdate) error
}

type BotChannelTracker struct {
	logger logging.Logger
}

func NewBotChannelTracker(logger logging.Logger) *BotChannelTracker {
	return &BotChannelTracker{logger: logger}
}

func (t *BotChannelTracker) GetBotVoiceState(guild *discordgo.Guild, session *discordgo.Session) *discordgo.VoiceState {
	if guild == nil || session == nil || session.State == nil || session.State.User == nil {
		t.logger.Warn("GetBotVoiceState recibió parámetros nulos",
			zap.Bool("guild_nil", guild == nil),
			zap.Bool("session_nil", session == nil),
		)
		return nil
	}
	for _, state := range guild.VoiceStates {
		if state.UserID == session.State.User.ID {
			return state
		}
	}
	return nil
}

func (t *BotChannelTracker) CountUsersInChannel(guild *discordgo.Guild, channelID string, excludeUserID string) int {
	if guild == nil {
		t.logger.Warn("CountUsersInChannel recibió guild nulo")
		return 0
	}
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

func (m *BotMover) MoveBotToNewChannel(ctx context.Context, guildPlayer ports.GuildPlayer, newChannelID string, oldChannelID string, userID string) error {
	logger := m.logger.With(
		zap.String("component", "BotMover"),
		zap.String("method", "MoveBotToNewChannel"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("oldChannel", oldChannelID),
		zap.String("newChannel", newChannelID),
		zap.String("userID", userID),
	)

	if err := guildPlayer.MoveToVoiceChannel(ctx, newChannelID); err != nil {
		logger.Error("Error al solicitar a GuildPlayer que se mueva al nuevo canal", zap.Error(err))
		return err
	}

	logger.Info("Solicitud de movimiento del bot al nuevo canal enviada a GuildPlayer exitosamente")
	return nil
}

type PlaybackController struct {
	logger logging.Logger
}

func NewPlaybackController(logger logging.Logger) *PlaybackController {
	return &PlaybackController{logger: logger}
}

func (c *PlaybackController) HandlePlayback(ctx context.Context, guildPlayer ports.GuildPlayer, currentSong *entity.PlayedSong, vs *discordgo.VoiceStateUpdate, botVoiceState *discordgo.VoiceState, guild *discordgo.Guild, session *discordgo.Session) error {
	logger := c.logger.With(
		zap.String("component", "PlaybackController"),
		zap.String("method", "HandlePlayback"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("guildID", vs.GuildID),
		zap.String("userID", vs.UserID),
	)

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
					logger.Info("Deteniendo reproducción por falta de usuarios en el canal",
						zap.String("channelID", botVoiceState.ChannelID),
						zap.String("track", currentSong.DiscordSong.TitleTrack))

					if err := guildPlayer.Stop(ctx); err != nil {
						logger.Error("Error al detener la reproducción", zap.Error(err))
						return fmt.Errorf("error al detener la reproducción: %w", err)
					}
				} else {
					logger.Debug("Usuarios aún presentes en el canal, continuando reproducción",
						zap.Int("usersCount", usersInBotChannel))
				}
			}
		}
	} else {
		logger.Debug("Cambio de estado de voz no relevante para la reproducción actual")
	}
	return nil
}

type VoiceStateService struct {
	tracker  *BotChannelTracker
	mover    *BotMover
	playback *PlaybackController
	logger   logging.Logger
}

func NewVoiceStateService(tracker *BotChannelTracker, mover *BotMover, playback *PlaybackController, logger logging.Logger) *VoiceStateService {
	return &VoiceStateService{
		tracker:  tracker,
		mover:    mover,
		playback: playback,
		logger:   logger,
	}
}

func (s *VoiceStateService) HandleVoiceStateChange(ctx context.Context, guildPlayer ports.GuildPlayer, session *discordgo.Session, vs *discordgo.VoiceStateUpdate) error {
	logger := s.logger.With(
		zap.String("component", "VoiceStateService"),
		zap.String("method", "HandleVoiceStateChange"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("guildID", vs.GuildID),
		zap.String("userID", vs.UserID),
	)

	guild, err := session.State.Guild(vs.GuildID)
	if err != nil {
		logger.Error("Error al obtener el servidor", zap.Error(err))
		return fmt.Errorf("error al obtener el servidor: %w", err)
	}

	botVoiceState := s.tracker.GetBotVoiceState(guild, session)
	logger.Debug("Estado de voz del bot obtenido",
		zap.Bool("botInVoice", botVoiceState != nil),
		zap.String("botChannel", getChannelID(botVoiceState)))

	if botVoiceState != nil {
		wasInBotChannel := vs.BeforeUpdate != nil && vs.BeforeUpdate.ChannelID == botVoiceState.ChannelID
		movedToDifferentChannel := vs.ChannelID != "" && vs.ChannelID != botVoiceState.ChannelID

		if wasInBotChannel && movedToDifferentChannel {
			newChannelID := vs.ChannelID
			usersInCurrentChannel := s.tracker.CountUsersInChannel(guild, botVoiceState.ChannelID, session.State.User.ID)
			logger.Debug("Usuario moviéndose desde el canal del bot",
				zap.Int("remainingUsers", usersInCurrentChannel),
				zap.String("newChannel", newChannelID))

			if usersInCurrentChannel == 0 {
				logger.Info("Moviendo bot a nuevo canal por falta de usuarios")
				return s.mover.MoveBotToNewChannel(ctx, guildPlayer, newChannelID, botVoiceState.ChannelID, vs.UserID)
			}
		}
	}

	currentSong, err := guildPlayer.GetPlayedSong(ctx)
	if err != nil {
		logger.Error("Error al obtener la canción actual", zap.Error(err))
		return fmt.Errorf("error al obtener la canción actual: %w", err)
	}

	var trackTitle string
	if currentSong != nil && currentSong.DiscordSong != nil {
		trackTitle = currentSong.DiscordSong.TitleTrack
	}

	logger.Debug("Canción actual obtenida",
		zap.Bool("hasCurrentSong", currentSong != nil),
		zap.String("track", trackTitle))

	return s.playback.HandlePlayback(ctx, guildPlayer, currentSong, vs, botVoiceState, guild, session)
}

func getChannelID(state *discordgo.VoiceState) string {
	if state == nil {
		return ""
	}
	return state.ChannelID
}
