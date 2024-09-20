package repository

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
)

// MetadataRepository define las operaciones para manejar metadatos.
type MetadataRepository interface {
	SaveMetadata(ctx context.Context, metadata model.Metadata) error
	GetMetadata(ctx context.Context, id string) (*model.Metadata, error)
	DeleteMetadata(ctx context.Context, id string) error
}
