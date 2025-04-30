package health

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/errors_app"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"time"
)

type DiscordChecker struct {
	session *discordgo.Session
	logger  logging.Logger
}

func NewDiscordChecker(session *discordgo.Session, logger logging.Logger) ports.DiscordHealthChecker {
	return &DiscordChecker{
		session: session,
		logger:  logger,
	}
}

func (d *DiscordChecker) Check(ctx context.Context) (entity.DiscordHealth, error) {
	ctx = trace.WithTraceID(ctx)
	traceID := trace.GetTraceID(ctx)

	logger := d.logger.With(
		zap.String("component", "DiscordChecker"),
		zap.String("method", "Check"),
		zap.String("trace_id", traceID),
	)

	logger.Debug("Iniciando verificación de salud de Discord")
	start := time.Now()
	health := entity.DiscordHealth{
		Connected: false,
	}

	defer func() {
		health.CheckDurationMS = float64(time.Since(start).Milliseconds())
	}()

	if d.session == nil {
		logger.Error("La sesión de Discord es nula")
		health.Error = "Sesión de Discord no inicializada"
		return health, errors_app.NewAppError(
			errors_app.ErrCodeInternalError,
			"Sesión de Discord no inicializada",
			nil,
		)
	}

	health.Connected = d.session.DataReady
	health.HeartbeatLatencyMS = d.session.HeartbeatLatency().Seconds() * 1000

	guilds := 0
	if d.session.State != nil {
		guilds = len(d.session.State.Guilds)
	}
	health.Guilds = guilds

	voiceConnections := 0
	if d.session.State != nil && d.session.State.User != nil {
		botID := d.session.State.User.ID
		voiceStates := make(map[string]string)

		for _, guild := range d.session.State.Guilds {
			if guild.VoiceStates != nil {
				for _, vs := range guild.VoiceStates {
					if vs.UserID == botID && vs.ChannelID != "" {
						voiceStates[guild.ID] = vs.ChannelID
						voiceConnections++
					}
				}
			}
		}
	}
	health.VoiceConnections = voiceConnections

	if d.session.State != nil {
		health.SessionID = d.session.State.SessionID
	}

	if !health.Connected {
		logger.Warn("Discord no está conectado")
		health.Error = "WebSocket no conectado"
	} else if health.HeartbeatLatencyMS > 1000 {
		logger.Warn("Discord tiene alta latencia",
			zap.Float64("heartbeat_latency_ms", health.HeartbeatLatencyMS),
		)
		health.Error = fmt.Sprintf("Alta latencia: %.2f ms", health.HeartbeatLatencyMS)
	} else {
		logger.Debug("Discord está saludable",
			zap.Float64("heartbeat_latency_ms", health.HeartbeatLatencyMS),
			zap.Int("guilds", health.Guilds),
			zap.Int("voice_connections", health.VoiceConnections),
		)
	}

	return health, nil
}
