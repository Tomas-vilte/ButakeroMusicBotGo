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
	mediaRepository ports.MediaRepository
}

func NewGetOperationStatusUseCase(mediaRepository ports.MediaRepository) *GetOperationStatusUseCaseImpl {
	return &GetOperationStatusUseCaseImpl{
		mediaRepository: mediaRepository,
	}
}

func (uc *GetOperationStatusUseCaseImpl) Execute(ctx context.Context, operationID, songID string) (*model.Media, error) {
	if !isValidUUID(operationID) {
		return nil, errors.ErrInvalidUUID.WithMessage(
			fmt.Sprintf("ID de operación inválido: %s", operationID))
	}

	operation, err := uc.mediaRepository.GetMedia(ctx, operationID, songID)
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
