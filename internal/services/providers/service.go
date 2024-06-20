package providers

import (
	"context"
	"google.golang.org/api/youtube/v3"
)

type (
	// YouTubeService define una interfaz para las operaciones relacionadas con YouTube.
	YouTubeService interface {
		SearchVideoID(ctx context.Context, searchTerm string) (string, error)
		GetVideoDetails(ctx context.Context, videoID string) (*youtube.Video, error)
	}
)
