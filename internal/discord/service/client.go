package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/internal/types"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

// ClientAPI implementa la interfaz APIClient
type ClientAPI struct {
	apiGatewayURL string
	logger        logging.Logger
	httpClient    *http.Client
}

// NewClient crea una nueva instancia de ClientAPI
func NewClient(apiGatewayURL string, logger logging.Logger) APIClient {
	return &ClientAPI{
		apiGatewayURL: apiGatewayURL,
		logger:        logger,
		httpClient:    &http.Client{Timeout: 10 * time.Second},
	}
}

// ProcessSongMetadata procesa los metadatos de una canción a través de API Gateway
func (c *ClientAPI) ProcessSongMetadata(ctx context.Context, input string) (*types.SongMetadata, error) {
	c.logger.Info("Procesando metadata de la cancion a traves de API Gateway", zap.String("input", input))

	requestBody := map[string]string{
		"key":  input,
		"song": input,
	}
	responseBody, err := c.makeRequest(ctx, requestBody)
	if err != nil {
		return nil, fmt.Errorf("error al procesar la metadata %w", err)
	}

	var metadata types.SongMetadata
	if err := json.Unmarshal(responseBody, &metadata); err != nil {
		c.logger.Error("Error al convertir el cuerpo de la respuesta a JSON", zap.Error(err))
	}

	return &metadata, nil
}

func (c *ClientAPI) makeRequest(ctx context.Context, requestBody interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		c.logger.Error("Error al convertir el cuerpo de la solicitud a JSON", zap.Error(err))
		return nil, fmt.Errorf("error al codificar el cuerpo de la solicitud: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiGatewayURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		c.logger.Error("Error al crear una nueva solicitud", zap.Error(err))
		return nil, fmt.Errorf("error al crear la solicitud: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("Error en la solicitud a API Gateway", zap.Error(err))
		return nil, fmt.Errorf("error en la solicitud a API Gateway: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error("Error en la solicitud a API Gateway", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("error en la solicitud a API Gateway: status code %d", resp.StatusCode)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("Error al leer el cuerpo de la respuesta", zap.Error(err))
		return nil, fmt.Errorf("error al leer la respuesta: %w", err)
	}

	return responseBody, nil
}
