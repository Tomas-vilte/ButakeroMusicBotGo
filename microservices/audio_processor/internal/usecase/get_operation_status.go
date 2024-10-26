package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/port"
	"github.com/google/uuid"
)

var (
	ErrInvalidUUID = errors.New("UUID inválido")
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
	// Validación de entrada: comprobar que los IDs son UUIDs válidos
	if !isValidUUID(operationID) || !isValidUUID(songID) {
		return nil, ErrInvalidUUID
	}

	operation, err := uc.operationRepository.GetOperationResult(ctx, operationID, songID)
	if err != nil {
		return &model.OperationResult{}, fmt.Errorf("error al obtener la operación: %w", err)
	}

	return operation, nil

}

func isValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
