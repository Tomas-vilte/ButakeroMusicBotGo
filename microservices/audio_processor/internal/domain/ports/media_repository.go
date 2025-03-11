package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
)

// MediaRepository define las operaciones para manejar registros de procesamiento multimedia.
type MediaRepository interface {
	// SaveMedia guarda un registro de procesamiento multimedia.
	SaveMedia(ctx context.Context, media *model.Media) error

	// GetMedia obtiene un registro de procesamiento multimedia por su ID y video_id.
	GetMedia(ctx context.Context, id, videoID string) (*model.Media, error)

	// DeleteMedia elimina un registro de procesamiento multimedia por su ID y video_id.
	DeleteMedia(ctx context.Context, id, videoID string) error

	// UpdateMedia actualiza el registro de procesamiento multimedia.
	UpdateMedia(ctx context.Context, id, videoID string, media *model.Media) error
}
