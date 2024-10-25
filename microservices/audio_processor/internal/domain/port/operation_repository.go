package port

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
)

// OperationRepository define las operaciones para manejar resultados de operación.
type OperationRepository interface {
	SaveOperationsResult(ctx context.Context, result model.OperationResult) error
	GetOperationResult(ctx context.Context, id, songID string) (*model.OperationResult, error)
	DeleteOperationResult(ctx context.Context, id, songID string) error
	UpdateOperationStatus(ctx context.Context, operationID string, songID string, status string) error
}
