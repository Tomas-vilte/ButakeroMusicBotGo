package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"time"
)

var (
	ErrEmptySongName  = errors.New("el nombre de la canción no puede estar vacío")
	ErrInvalidBaseURL = errors.New("la URL base no es válida")
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

	return &AudioAPIClient{
		baseURL: baseURL,
		logger:  logger,
		httpClient: &http.Client{
			Timeout:   config.Timeout,
			Transport: transport,
		},
	}, nil
}

func (c *AudioAPIClient) DownloadSong(ctx context.Context, songName string) (*entity.DownloadResponse, error) {
	if songName == "" {
		return nil, ErrEmptySongName
	}

	endpoint := c.baseURL.JoinPath("api/v1/audio/start")

	params := url.Values{}
	params.Add("song", songName)
	endpoint.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("hubo un error al crear la request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("falló la request: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Error("No se pudo cerrar el body de la respuesta",
				zap.Error(closeErr),
				zap.String("songName", songName))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("el código de estado %d no es el esperado, debería ser %d",
			resp.StatusCode, http.StatusOK)
	}

	var response entity.DownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("no se pudo decodificar la respuesta: %w", err)
	}

	c.logger.Info("Iniciando descarga",
		zap.String("songName", songName),
		zap.String("operationId", response.OperationID),
		zap.String("songId", response.SongID))

	return &response, nil
}
