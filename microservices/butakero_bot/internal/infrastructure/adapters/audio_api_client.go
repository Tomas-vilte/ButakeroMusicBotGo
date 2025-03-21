package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/errors_app"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"time"
)

type AudioAPIClientConfig struct {
	BaseURL         string
	Timeout         time.Duration
	MaxIdleConns    int
	MaxConnsPerHost int
}

type AudioAPIClient struct {
	baseURL    *url.URL
	logger     logging.Logger
	httpClient *http.Client
}

func NewAudioAPIClient(config AudioAPIClientConfig, logger logging.Logger) (*AudioAPIClient, error) {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	baseURL, err := url.Parse(config.BaseURL)
	if err != nil {
		return nil, errors_app.NewAppError(errors_app.ErrCodeInvalidInput, "La URL base no es válida", err)
	}

	transport := &http.Transport{
		MaxIdleConns:    config.MaxIdleConns,
		MaxConnsPerHost: config.MaxConnsPerHost,
		IdleConnTimeout: 90 * time.Second,
	}

	logger = logger.With(
		zap.String("component", "audio_api_client"),
		zap.String("baseURL", config.BaseURL),
	)

	return &AudioAPIClient{
		baseURL: baseURL,
		logger:  logger,
		httpClient: &http.Client{
			Timeout:   config.Timeout,
			Transport: transport,
		},
	}, nil
}

func (c *AudioAPIClient) DownloadSong(ctx context.Context, songName, providerType string) (*entity.DownloadResponse, error) {
	logger := c.logger.With(
		zap.String("method", "DownloadSong"),
		zap.String("songName", songName),
		zap.String("providerType", providerType),
	)

	if songName == "" || providerType == "" {
		logger.Error("Parámetros vacíos")
		return nil, errors_app.NewAppError(errors_app.ErrCodeInvalidInput, "Faltan algunos parámetros provider_type/song", nil)
	}

	endpoint := c.baseURL.JoinPath("api/v1/audio/start")

	params := url.Values{}
	params.Add("song", songName)
	params.Add("provider_type", providerType)
	endpoint.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), nil)
	if err != nil {
		logger.Error("Error al crear la solicitud", zap.Error(err))
		return nil, errors_app.NewAppError(errors_app.ErrCodeInternalError, "Hubo un error al crear la solicitud", err)
	}

	logger.Debug("Enviando solicitud de descarga", zap.String("endpoint", endpoint.String()))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Error al realizar la solicitud", zap.Error(err))
		return nil, errors_app.NewAppError(errors_app.ErrCodeYouTubeAPIError, "El servicio de música no está disponible en este momento", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Error("Error al cerrar el body de la respuesta", zap.Error(closeErr))
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error al leer el cuerpo de la respuesta", zap.Error(err))
		return nil, errors_app.NewAppError(errors_app.ErrCodeInternalError, "Error al leer el cuerpo de la respuesta", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiError entity.APIError
		if err := json.Unmarshal(body, &apiError); err != nil {
			logger.Error("Error al decodificar el error de API",
				zap.Error(err),
				zap.Int("statusCode", resp.StatusCode),
				zap.String("responseBody", string(body)),
			)
			return nil, errors_app.NewAppError(errors_app.ErrCodeInternalError, fmt.Sprintf("Error en la solicitud (Código: %d)", resp.StatusCode), err)
		}

		apiErr := errors_app.GetAPIError(apiError.Error.Code)
		logger.Error("Error de API detectado",
			zap.String("code", string(apiErr.Code)),
			zap.String("message", apiErr.Message),
			zap.String("originalCode", apiError.Error.Code),
			zap.String("originalMessage", apiError.Error.Message),
		)
		return nil, apiErr
	}

	var response entity.DownloadResponse
	if err := json.Unmarshal(body, &response); err != nil {
		logger.Error("Error al decodificar la respuesta", zap.Error(err))
		return nil, errors_app.NewAppError(errors_app.ErrCodeInternalError, "No se pudo decodificar la respuesta", err)
	}

	logger.Info("Iniciando descarga",
		zap.String("videoID", response.VideoID),
		zap.String("status", response.Status),
	)

	return &response, nil
}
