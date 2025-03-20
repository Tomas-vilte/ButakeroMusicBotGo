package usecase

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
)

type GetOperationStatusUseCaseImpl struct {
	mediaRepository ports.MediaRepository
}

func NewGetOperationStatusUseCase(mediaRepository ports.MediaRepository) *GetOperationStatusUseCaseImpl {
	return &GetOperationStatusUseCaseImpl{
		mediaRepository: mediaRepository,
	}
}

func (uc *GetOperationStatusUseCaseImpl) Execute(ctx context.Context, videoID string) (*model.Media, error) {
	if videoID == "" {
		return nil, errors.ErrInvalidInput.WithMessage("El ID de la canción no puede estar vacío")
	}

	operation, err := uc.mediaRepository.GetMedia(ctx, videoID)
	if err != nil {
		return nil, err
	}
	return operation, nil
}
