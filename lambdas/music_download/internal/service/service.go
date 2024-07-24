package service

import (
	"context"
	"google.golang.org/api/youtube/v3"
)

// YouTubeService define una interfaz para las operaciones relacionadas con YouTube.
type YouTubeService interface {
	SearchVideoID(ctx context.Context, searchTerm string) (string, error)
	GetVideoDetails(ctx context.Context, videoID string) (*youtube.Video, error)
}
