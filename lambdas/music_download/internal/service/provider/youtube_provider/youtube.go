package youtube_provider

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/logging"
	"go.uber.org/zap"
	"google.golang.org/api/youtube/v3"
)

type (
	// YouTubeClient define una interfaz para las operaciones con la API de YouTube.
	YouTubeClient interface {
		VideosListCall(ctx context.Context, part []string) VideosListCallWrapper
		SearchListCall(ctx context.Context, part []string) SearchListCallWrapper
	}

	// YouTubeProvider implementa la interface providers.Service
	YouTubeProvider struct {
		logger logging.Logger
		Client YouTubeClient
	}
)

func NewYouTubeProvider(logger logging.Logger, client YouTubeClient) *YouTubeProvider {
	return &YouTubeProvider{
		logger: logger,
		Client: client,
	}
}

// SearchVideoID Busca el ID de un video en YouTube por un termino de busqueda
func (p *YouTubeProvider) SearchVideoID(ctx context.Context, searchTerm string) (string, error) {
	p.logger.Info("Buscando video en YouTube", zap.String("searchTerm", searchTerm))
	call := p.Client.SearchListCall(ctx, []string{"id"}).Q(searchTerm).MaxResults(1).Type("video")

	response, err := call.Do()
	if err != nil {
		p.logger.Error("Error al buscar vídeo en YouTube", zap.Error(err))
		return "", fmt.Errorf("error al buscar vídeo en YouTube: %w", err)
	}

	if len(response.Items) == 0 {
		p.logger.Info("No se encontró ningún vídeo para el término de búsqueda", zap.String("searchTerm", searchTerm))
		return "", fmt.Errorf("no se encontró ningún vídeo para el término de búsqueda: %s", searchTerm)
	}
	videoID := response.Items[0].Id.VideoId
	p.logger.Info("Video encontrado", zap.String("videoID", videoID))
	return videoID, nil
}

// GetVideoDetails obtiene los detalles de un video de YouTube por su ID.
func (p *YouTubeProvider) GetVideoDetails(ctx context.Context, videoID string) (*youtube.Video, error) {
	p.logger.Info("Obteniendo detalles del video desde Youtube", zap.String("videoID", videoID))
	call := p.Client.VideosListCall(ctx, []string{"snippet", "contentDetails", "liveStreamingDetails"}).Id(videoID)
	response, err := call.Do()
	if err != nil {
		p.logger.Error("Error al obtener detalles del video desde Youtube", zap.Error(err))
		return nil, fmt.Errorf("error al obtener detalles del video: %w", err)
	}

	if len(response.Items) == 0 {
		p.logger.Info("Video no encontrado", zap.String("videoID", videoID))
		return nil, fmt.Errorf("video no encontrado con el ID: %s", videoID)
	}

	p.logger.Info("Detalles del video recuperados exitosamente",
		zap.String("videoID", videoID),
		zap.String("title", response.Items[0].Snippet.Title))

	return response.Items[0], nil
}
