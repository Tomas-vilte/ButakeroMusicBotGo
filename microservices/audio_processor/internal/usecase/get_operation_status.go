package usecase

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/google/uuid"
)

type GetOperationStatusUseCaseImpl struct {
	operationRepository ports.OperationRepository
}

func NewGetOperationStatusUseCase(operationRepository ports.OperationRepository) *GetOperationStatusUseCaseImpl {
	return &GetOperationStatusUseCaseImpl{
		operationRepository: operationRepository,
	}
}

func (uc *GetOperationStatusUseCaseImpl) Execute(ctx context.Context, operationID, songID string) (*model.OperationResult, error) {
	if !isValidUUID(operationID) {
		return nil, errors.ErrInvalidUUID.WithMessage(
			fmt.Sprintf("ID de operación inválido: %s", operationID))
	}

	operation, err := uc.operationRepository.GetOperationResult(ctx, operationID, songID)
	if err != nil {
		return nil, errors.ErrOperationNotFound.WithMessage(
			fmt.Sprintf("Operación no encontrada: %s", operationID))
	}
	return operation, nil
}

func isValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
