package usecase

import (
	"context"
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

func (uc *GetOperationStatusUseCaseImpl) Execute(ctx context.Context, operationID, videoID string) (*model.Media, error) {
	if !isValidUUID(operationID) {
		return nil, errors.ErrInvalidUUID.WithMessage("UUID inválido")
	}

	if videoID == "" {
		return nil, errors.ErrInvalidInput.WithMessage("El ID de la canción no puede estar vacío")
	}

	operation, err := uc.mediaRepository.GetMedia(ctx, operationID, videoID)
	if err != nil {
		return nil, errors.ErrOperationNotFound.WithMessage("No se encontró la operación solicitada")
	}
	return operation, nil
}

func isValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
