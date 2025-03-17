package adapters

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	ErrEmptyParameters = errors.New("faltan algunos parametros provider_type/songName")
	ErrInvalidBaseURL  = errors.New("la URL base no es válida")
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
		return nil, fmt.Errorf("%w: %v", ErrInvalidBaseURL, err)
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
		logger.Error("Parámetros vacíos", zap.Error(ErrEmptyParameters))
		return nil, ErrEmptyParameters
	}

	endpoint := c.baseURL.JoinPath("api/v1/audio/start")

	params := url.Values{}
	params.Add("song", songName)
	params.Add("provider_type", providerType)
	endpoint.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), nil)
	if err != nil {
		logger.Error("Error al crear la solicitud", zap.Error(err))
		return nil, fmt.Errorf("hubo un error al crear la request: %w", err)
	}

	logger.Debug("Enviando solicitud de descarga", zap.String("endpoint", endpoint.String()))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Error al realizar la solicitud", zap.Error(err))
		return nil, fmt.Errorf("falló la request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Error("Error al cerrar el body de la respuesta", zap.Error(closeErr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.Error("Error en la solicitud",
			zap.Int("statusCode", resp.StatusCode),
			zap.String("responseBody", string(body)),
		)
		return nil, fmt.Errorf("error en la solicitud: código de estado %d, respuesta: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error al leer el cuerpo de la respuesta", zap.Error(err))
		return nil, fmt.Errorf("error al leer el cuerpo de la respuesta: %w", err)
	}

	var response entity.DownloadResponse
	if err := json.Unmarshal(body, &response); err != nil {
		logger.Error("Error al decodificar la respuesta", zap.Error(err))
		return nil, fmt.Errorf("no se pudo decodificar la respuesta: %w", err)
	}

	logger.Info("Iniciando descarga",
		zap.String("videoID", response.VideoID),
		zap.String("status", response.Status),
	)

	return &response, nil
}
