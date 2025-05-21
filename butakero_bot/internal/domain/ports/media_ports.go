package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model"
)

// MediaClient es una interfaz que define los métodos para interactuar con un cliente de medias mediante API.
type MediaClient interface {
	// GetMediaByID obtiene un medio por su ID.
	GetMediaByID(ctx context.Context, videoID string) (*model.Media, error)
	// GetMediaByURL obtiene un medio por su URL.
	SearchMediaByTitle(ctx context.Context, title string) ([]*model.Media, error)
}
