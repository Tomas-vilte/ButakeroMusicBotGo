package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
)

// VideoProvider define la interfaz para interactuar con servicios de video como YouTube o Spotify entre otros.
type VideoProvider interface {
	// GetVideoDetails obtiene los detalles del video usando su ID.
	GetVideoDetails(ctx context.Context, videoID string) (*model.MediaDetails, error)
	// SearchVideoID busca el ID del primer video que coincida con la consulta dada.
	SearchVideoID(ctx context.Context, input string) (string, error)
}
