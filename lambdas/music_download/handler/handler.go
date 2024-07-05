package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/downloader"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/logging"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/uploader"
	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
	"net/http"
)

// SongEvent representa la estructura del evento de la canción
type SongEvent struct {
	URL string `json:"url"`
	Key string `json:"key"`
}

// EventManager define una interfaz para manejar eventos Lambda
type EventManager interface {
	HandleEvent(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
}

// Handler es la estructura que maneja los eventos Lambda
type Handler struct {
	Downloader downloader.Downloader
	Uploader   uploader.Uploader
	Logger     logging.Logger
}

// NewHandler crea un nuevo Handler con los componentes necesarios
func NewHandler(downloader downloader.Downloader, uploader uploader.Uploader, logger logging.Logger) *Handler {
	return &Handler{
		Downloader: downloader,
		Uploader:   uploader,
		Logger:     logger,
	}
}

// HandleEvent maneja el evento Lambda
func (h *Handler) HandleEvent(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.Logger.Info("Evento recibido", zap.Any("event", event))

	var songEvent SongEvent
	err := json.Unmarshal([]byte(event.Body), &songEvent)
	if err != nil {
		h.Logger.Error("Error al parsear el evento", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf("Error al parsear el evento: %v", err),
		}, fmt.Errorf("error al parsear el evento: %v", err)
	}

	key := fmt.Sprintf("audio_input_raw/%s.m4a", songEvent.Key)
	err = h.Downloader.DownloadSong(songEvent.URL, key)
	if err != nil {
		h.Logger.Error("Error al descargar la canción", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf("Error al descargar la canción: %v", err),
		}, fmt.Errorf("error al descargar la canción: %v", err)
	}

	h.Logger.Info("Canción procesada exitosamente", zap.String("url", songEvent.URL), zap.String("key", songEvent.Key))
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "Canción procesada exitosamente",
	}, nil
}
