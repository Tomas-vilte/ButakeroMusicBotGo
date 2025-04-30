package health

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/errors_app"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultTimeout     = 10 * time.Second
	healthCheckTimeout = 5 * time.Second
	idleConnTimeout    = 90 * time.Second
	healthEndpointPath = "api/v1/health"
)

type ServiceConfig struct {
	BaseURL         string
	Timeout         time.Duration
	MaxIdleConns    int
	MaxConnsPerHost int
}

type ServiceBChecker struct {
	client  *http.Client
	baseURL *url.URL
	logger  logging.Logger
}

func NewServiceBChecker(config *ServiceConfig, logger logging.Logger) (ports.ServiceBHealthChecker, error) {
	if config == nil {
		return nil, errors_app.NewAppError(errors_app.ErrCodeInvalidInput, "La configuración del servicio no puede ser nula", nil)
	}

	if config.BaseURL == "" {
		return nil, errors_app.NewAppError(errors_app.ErrCodeInvalidInput, "La URL base no puede estar vacía", nil)
	}

	timeout := config.Timeout
	if timeout == 0 {
		timeout = defaultTimeout
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, errors_app.NewAppError(errors_app.ErrCodeInvalidInput, "La URL base no es válida", err)
	}

	maxIdleConns := config.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = 10
	}

	maxConnsPerHost := config.MaxConnsPerHost
	if maxConnsPerHost <= 0 {
		maxConnsPerHost = 10
	}

	transport := &http.Transport{
		MaxIdleConns:    maxIdleConns,
		MaxConnsPerHost: maxConnsPerHost,
		IdleConnTimeout: idleConnTimeout,
	}

	return &ServiceBChecker{
		baseURL: baseURL,
		logger:  logger,
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
	}, nil
}

func (s *ServiceBChecker) Check(ctx context.Context) (entity.ServiceBHealth, error) {
	ctx = trace.WithTraceID(ctx)
	traceID := trace.GetTraceID(ctx)

	logger := s.logger.With(
		zap.String("component", "ServiceBChecker"),
		zap.String("method", "Check"),
		zap.String("trace_id", traceID),
	)

	logger.Debug("Iniciando verificación de salud de Service B")
	start := time.Now()

	endpoint := s.baseURL.JoinPath(healthEndpointPath)

	ctx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		logger.Error("Error al crear la petición HTTP",
			zap.String("endpoint", endpoint.String()),
			zap.Error(err),
		)
		return entity.ServiceBHealth{}, errors_app.NewAppError(
			errors_app.ErrCodeInternalError,
			"Error al crear la petición HTTP",
			err,
		)
	}

	resp, err := s.client.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		logger.Warn("Error al conectar con Service B",
			zap.String("endpoint", endpoint.String()),
			zap.Int64("latency_ms", latency),
			zap.Error(err),
		)
		return entity.ServiceBHealth{
			Connected: false,
			LatencyMS: int(latency),
			Error:     fmt.Sprintf("Error de conexión: %v", err),
		}, nil
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logger.Error("Error al cerrar el cuerpo de la respuesta", zap.Error(err))
		}
	}()

	connected := resp.StatusCode == http.StatusOK
	statusText := fmt.Sprintf("HTTP %d", resp.StatusCode)

	health := entity.ServiceBHealth{
		Connected: connected,
		LatencyMS: int(latency),
		Status:    statusText,
	}

	if !connected {
		logger.Warn("Service B devolvió un estado no saludable",
			zap.Int("status_code", resp.StatusCode),
			zap.Int64("latency_ms", latency),
		)
	} else {
		logger.Debug("Service B está saludable",
			zap.Int("status_code", resp.StatusCode),
			zap.Int64("latency_ms", latency),
		)
	}
	return health, nil
}
