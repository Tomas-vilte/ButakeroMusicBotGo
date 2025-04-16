package api

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model"
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

type MediaAPIClient struct {
	baseURL    *url.URL
	logger     logging.Logger
	httpClient *http.Client
}

func NewMediaAPIClient(config AudioAPIClientConfig, logger logging.Logger) (*MediaAPIClient, error) {
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

	return &MediaAPIClient{
		baseURL: baseURL,
		logger:  logger,
		httpClient: &http.Client{
			Timeout:   config.Timeout,
			Transport: transport,
		},
	}, nil
}

func (c *MediaAPIClient) GetMediaByID(ctx context.Context, id string) (*model.Media, error) {
	logger := c.logger.With(
		zap.String("component", "MediaAPIClient"),
		zap.String("method", "GetMediaByID"),
		zap.String("id", id),
	)

	endpoint := c.baseURL.JoinPath("api/v1/media")

	params := url.Values{}
	params.Add("video_id", id)
	endpoint.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		logger.Error("Error al crear la solicitud", zap.Error(err))
		return nil, errors_app.NewAppError(errors_app.ErrCodeInternalError, "Hubo un error al crear la solicitud", err)
	}

	logger.Debug("Consultando cancion", zap.String("endpoint", endpoint.String()))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Error al realizar la solicitud", zap.Error(err))
		return nil, errors_app.NewAppError(errors_app.ErrCodeInternalError, "Hubo realizar la solicitud", err)
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
		var apiError model.ErrorResponse
		if err := json.Unmarshal(body, &apiError); err != nil {
			logger.Error("Error al decodificar el error de API",
				zap.Error(err),
				zap.Int("statusCode", resp.StatusCode),
				zap.String("responseBody", string(body)),
			)
			return nil, errors_app.NewAppError(errors_app.ErrCodeInternalError, fmt.Sprintf("Error en la solicitud (Código: %d)", resp.StatusCode), err)
		}
	}

	var response model.MediaResponse
	if err := json.Unmarshal(body, &response); err != nil {
		logger.Error("Error al decodificar la respuesta", zap.Error(err))
		return nil, errors_app.NewAppError(errors_app.ErrCodeInternalError, "No se pudo decodificar la respuesta", err)
	}

	if !response.Success {
		return nil, errors_app.NewAppError(errors_app.ErrCodeInternalError, "No se pudo obtener la canción", fmt.Errorf("error al obtener el media con el id: %s", id))
	}

	logger.Info("Cancion encontrada con exito",
		zap.String("id", id),
		zap.String("title", response.Data.Metadata.Title),
	)

	return response.Data, nil

}

func (c *MediaAPIClient) SearchMediaByTitle(ctx context.Context, title string) ([]*model.Media, error) {
	logger := c.logger.With(
		zap.String("component", "MediaAPIClient"),
		zap.String("method", "SearchMediaByTitle"),
		zap.String("title", title),
	)

	endpoint := c.baseURL.JoinPath("api/v1/media/search")

	params := url.Values{}
	params.Add("title", title)
	endpoint.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		logger.Error("Error al crear la solicitud", zap.Error(err))
		return nil, errors_app.NewAppError(errors_app.ErrCodeInternalError, "Hubo un error al crear la solicitud", err)
	}

	logger.Debug("Consultando cancion", zap.String("endpoint", endpoint.String()))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Error al realizar la solicitud", zap.Error(err))
		return nil, errors_app.NewAppError(errors_app.ErrCodeInternalError, "Hubo realizar la solicitud", err)
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
		return nil, errors_app.NewAppError(errors_app.ErrCodeInternalError, fmt.Sprintf("Error en la solicitud (Código: %d)", resp.StatusCode), err)
	}

	var response model.MediaListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		logger.Error("Error al decodificar la respuesta", zap.Error(err))
		return nil, errors_app.NewAppError(errors_app.ErrCodeInternalError, "No se pudo decodificar la respuesta", err)
	}

	if !response.Success {
		return nil, errors_app.NewAppError(errors_app.ErrCodeInternalError, "No se pudo obtener la canción", fmt.Errorf("error al obtener la cancion con el titulo: %s", title))
	}

	logger.Info("Cancion con el titulo encontrado con exito",
		zap.String("title", response.Data[0].Metadata.Title),
	)

	return response.Data, nil
}
