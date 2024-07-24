package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/api/youtube_api"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/cache"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/downloader"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/uploader"
	"github.com/aws/aws-lambda-go/events"
	"go.uber.org/zap"
	"net/http"
)

// SongEvent representa la estructura del evento de la canción
type SongEvent struct {
	Song string `json:"song"`
	Key  string `json:"key"`
}

// EventManager define una interfaz para manejar eventos Lambda
type EventManager interface {
	HandleEvent(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)
}

// Handler es la estructura que maneja los eventos Lambda
type Handler struct {
	downloader    downloader.Downloader
	cache         cache.Cache
	youTubeClient youtube_api.SongLooker
	uploader      uploader.Uploader
	logger        logging.Logger
}

// NewHandler crea un nuevo Handler con los componentes necesarios
func NewHandler(downloader downloader.Downloader, uploader uploader.Uploader, logger logging.Logger, clientYouTube youtube_api.SongLooker, cache cache.Cache) *Handler {
	return &Handler{
		downloader:    downloader,
		uploader:      uploader,
		logger:        logger,
		youTubeClient: clientYouTube,
		cache:         cache,
	}
}

// HandleEvent maneja el evento Lambda
func (h *Handler) HandleEvent(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	h.logger.Info("Evento recibido", zap.Any("event", event))

	var songEvent SongEvent
	err := json.Unmarshal([]byte(event.Body), &songEvent)
	if err != nil {
		h.logger.Error("Error al parsear el evento", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf("Error al parsear el evento: %v", err),
		}, fmt.Errorf("error al parsear el evento: %v", err)
	}

	cachedSong, err := h.cache.GetSong(ctx, songEvent.Key)
	if err != nil {
		h.logger.Error("Error al obtener del cache", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf("Error al obtener del cache: %v", err),
		}, fmt.Errorf("error al obtener del cache: %v", err)
	}
	if cachedSong != nil {
		h.logger.Info("Canción encontrada en cache", zap.String("song", songEvent.Song), zap.String("key", songEvent.Key))
		songDetails, err := json.Marshal(cachedSong)
		if err != nil {
			h.logger.Error("Error al serializar los detalles de la canción", zap.Error(err))
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusInternalServerError,
				Body:       fmt.Sprintf("Error al serializar los detalles de la canción: %v", err),
			}, fmt.Errorf("error al serializar los detalles de la canción: %v", err)
		}
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Body:       string(songDetails),
		}, nil
	}

	videoID, err := h.youTubeClient.SearchYouTubeVideoID(ctx, songEvent.Song)
	if err != nil {
		h.logger.Error("Error al buscar el ID del video en YouTube", zap.Error(err), zap.String("input", songEvent.Song))
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf("Error al buscar el ID del video: %v", err),
		}, fmt.Errorf("error al buscar el ID del video en YouTube: %v", err)
	}

	songs, err := h.youTubeClient.LookupSongs(ctx, videoID)
	if err != nil {
		h.logger.Error("Error al obtener detalles del video", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf("Error al obtener detalles del video: %v", err),
		}, fmt.Errorf("error al obtener detalles del video")
	}

	if len(songs) == 0 {
		h.logger.Error("No se encontraron detalles del video")
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "No se encontraron detalles del video",
		}, nil
	}

	key := fmt.Sprintf("audio_input_raw/%s.m4a", songEvent.Key)
	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	err = h.downloader.DownloadSong(videoURL, key)
	if err != nil {
		h.logger.Error("Error al descargar la canción", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf("Error al descargar la canción: %v", err),
		}, fmt.Errorf("error al descargar la canción: %v", err)
	}

	h.logger.Info("Canción procesada exitosamente", zap.String("song", songEvent.Song), zap.String("key", songEvent.Key))

	err = h.cache.SetSong(ctx, songEvent.Key, songs[0])
	if err != nil {
		h.logger.Error("Error al guardar en cache", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf("Error al guardar en cache: %v", err),
		}, fmt.Errorf("error al guardar en cache: %v", err)
	}

	songDetails, err := json.Marshal(songs[0])
	if err != nil {
		h.logger.Error("Error al serializar los detalles de la canción", zap.Error(err))
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf("Error al serializar los detalles de la canción: %v", err),
		}, fmt.Errorf("error al serializar los detalles de la canción: %v", err)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       string(songDetails),
	}, nil
}
