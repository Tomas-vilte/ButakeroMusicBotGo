package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model"
)

type MediaClient interface {
	GetMediaByID(ctx context.Context, videoID string) (*model.Media, error)
	SearchMediaByTitle(ctx context.Context, title string) ([]*model.Media, error)
}
