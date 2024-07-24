package youtube_api

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/service"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/types"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

type (
	SongLooker interface {
		LookupSongs(ctx context.Context, input string) ([]*types.Song, error)
		SearchYouTubeVideoID(ctx context.Context, searchTerm string) (string, error)
	}

	YouTubeFetcher struct {
		logger         logging.Logger
		youtubeService service.YouTubeService
	}
)

func NewYoutubeFetcher(logger logging.Logger, youtubeService service.YouTubeService) *YouTubeFetcher {
	return &YouTubeFetcher{
		logger:         logger,
		youtubeService: youtubeService,
	}
}

func (yt *YouTubeFetcher) LookupSongs(ctx context.Context, input string) ([]*types.Song, error) {
	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", input)

	video, err := yt.youtubeService.GetVideoDetails(ctx, input)
	if err != nil {
		yt.logger.Error("Error al obtener detalles del video", zap.Error(err))
		return nil, fmt.Errorf("error al obtener detalles del video")
	}

	duration, err := parseCustomDuration(video.ContentDetails.Duration)
	if err != nil {
		yt.logger.Error("Error al analizar la duracion: ", zap.Error(err))
		return nil, fmt.Errorf("error al analizar la duracion")
	}
	thumbnailURL := video.Snippet.Thumbnails.Default.Url

	song := &types.Song{
		Type:         "youtube_provider",
		Title:        video.Snippet.Title,
		URL:          videoURL,
		Playable:     video.Snippet.LiveBroadcastContent != "live",
		ThumbnailURL: &thumbnailURL,
		Duration:     duration,
	}
	return []*types.Song{song}, nil
}

func (yt *YouTubeFetcher) SearchYouTubeVideoID(ctx context.Context, searchTerm string) (string, error) {
	videoID, err := yt.youtubeService.SearchVideoID(ctx, searchTerm)
	if err != nil {
		yt.logger.Error("Error al buscar el video en Youtube", zap.Error(err))
		return "", fmt.Errorf("error al buscar el video en YouTube: %w", err)
	}
	return videoID, nil
}

func parseCustomDuration(durationStr string) (time.Duration, error) {
	parts := strings.Split(durationStr, "T")
	if len(parts) != 2 {
		return 0, fmt.Errorf("formato de duración no válido: %s", durationStr)
	}

	durationPart := parts[1]
	var duration time.Duration

	for durationPart != "" {
		var value int64
		var unit string

		// Obtener el número y la unidad
		i := 0
		for ; i < len(durationPart); i++ {
			if durationPart[i] < '0' || durationPart[i] > '9' {
				break
			}
		}

		value, err := strconv.ParseInt(durationPart[:i], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("formato de duración no válido: %s", durationStr)
		}

		unit = durationPart[i : i+1]

		// Actualizar la duración
		switch unit {
		case "H":
			duration += time.Duration(value) * time.Hour
		case "M":
			duration += time.Duration(value) * time.Minute
		case "S":
			duration += time.Duration(value) * time.Second
		default:
			return 0, fmt.Errorf("unidad de duración desconocida: %s", unit)
		}

		// Mover al siguiente componente de la duración
		durationPart = durationPart[i+1:]
	}

	return duration, nil
}
