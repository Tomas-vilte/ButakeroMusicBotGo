package api

import (
	"encoding/json"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const (
	healthCheckContextKey  = "health_check"
	healthCheckContentType = "application/json"
)

type HealthHandler struct {
	discordChecker  ports.DiscordHealthChecker
	serviceBChecker ports.ServiceBHealthChecker
	logger          logging.Logger
	cfg             *config.Config
}

func NewHealthHandler(
	discordChecker ports.DiscordHealthChecker,
	serviceBChecker ports.ServiceBHealthChecker,
	logger logging.Logger,
	cfg *config.Config,
) *HealthHandler {
	return &HealthHandler{
		discordChecker:  discordChecker,
		serviceBChecker: serviceBChecker,
		logger:          logger,
		cfg:             cfg,
	}
}

func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := trace.WithTraceID(r.Context())
	traceID := trace.GetTraceID(ctx)

	logger := h.logger.With(
		zap.String("component", "HealthHandler"),
		zap.String("method", "Handle"),
		zap.String("trace_id", traceID),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
	)

	logger.Debug("Iniciando verificación de salud del sistema")

	response := entity.HealthResponse{
		Status:    entity.StatusOperational,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   h.cfg.AppVersion,
	}

	discordHealth, err := h.discordChecker.Check(ctx)
	if err != nil {
		logger.Warn("Error al verificar la salud de Discord",
			zap.Error(err),
		)
		response.Discord = entity.DiscordHealth{
			Connected: false,
			Error:     err.Error(),
		}
	} else {
		response.Discord = discordHealth
	}

	serviceBHealth, err := h.serviceBChecker.Check(ctx)
	if err != nil {
		logger.Warn("Error al verificar la salud de Service B",
			zap.Error(err),
		)
		response.ServiceB = entity.ServiceBHealth{
			Connected: false,
			Error:     err.Error(),
		}
	} else {
		response.ServiceB = serviceBHealth
	}

	response.Status, response.Message = entity.DetermineOverallStatus(response.Discord, response.ServiceB)

	logger.Debug("Verificación de salud completada",
		zap.String("status", string(response.Status)),
		zap.String("message", response.Message),
		zap.Bool("discord_connected", response.Discord.Connected),
		zap.Bool("service_b_connected", response.ServiceB.Connected),
		zap.Float64("duration_ms", float64(time.Since(start).Milliseconds())),
	)

	statusCode := http.StatusOK
	switch response.Status {
	case entity.StatusDegraded:
		statusCode = http.StatusOK
	case entity.StatusDown:
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", healthCheckContentType)
	w.Header().Set("X-Trace-ID", traceID)
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("Error al codificar la respuesta de salud",
			zap.Error(err),
		)
		http.Error(w, "Fallo al codificar la respuesta de salud", http.StatusInternalServerError)
		return
	}
}
