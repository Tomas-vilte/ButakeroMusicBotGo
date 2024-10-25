package usecase

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/port"
)

type GetOperationStatusUseCaseImpl struct {
	operationRepository port.OperationRepository
}

func NewGetOperationStatusUseCase(operationRepository port.OperationRepository) *GetOperationStatusUseCaseImpl {
	return &GetOperationStatusUseCaseImpl{
		operationRepository: operationRepository,
	}
}

func (uc *GetOperationStatusUseCaseImpl) Execute(ctx context.Context, operationID, songID string) (*model.OperationResult, error) {
	operation, err := uc.operationRepository.GetOperationResult(ctx, operationID, songID)
	if err != nil {
		return &model.OperationResult{}, fmt.Errorf("error al obtener la operación: %w", err)
	}

	return operation, nil

}
