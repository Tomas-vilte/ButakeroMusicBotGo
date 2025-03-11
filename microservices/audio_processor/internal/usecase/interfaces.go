package usecase

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
)

type (
	InitialDownloadUseCase interface {
		Execute(ctx context.Context, song string, providerType string) (*model.OperationInitResult, error)
	}

	GetOperationStatusUseCase interface {
		Execute(ctx context.Context, operationID, songID string) (*model.Media, error)
	}
)
