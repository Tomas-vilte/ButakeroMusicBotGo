package service

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/internal/types"
)

type (
	// APIClient esta interface interactua con la API Gateway
	APIClient interface {
		ProcessSongMetadata(ctx context.Context, input string) (*types.SongMetadata, error)
	}
)
