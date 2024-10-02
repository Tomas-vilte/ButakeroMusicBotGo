package usecase

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
)

type (
	InitialDownloadUseCase interface {
		Execute(ctx context.Context, song string) (string, string, error)
	}

	GetOperationStatusUseCase interface {
		Execute(ctx context.Context, operationID, songID string) (model.OperationResult, error)
	}
)
