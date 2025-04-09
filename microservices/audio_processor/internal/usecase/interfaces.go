package usecase

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
)

type (
	GetOperationStatusUseCase interface {
		Execute(ctx context.Context, videoID string) (*model.Media, error)
	}
)
