package usecase

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	errorsApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
)

type InitiateDownloadUseCase struct {
	coreService      ports.CoreService
	providerService  ports.VideoService
	operationService ports.OperationService
}

func NewInitiateDownloadUseCase(coreService ports.CoreService, providerAPI ports.VideoService, operationService ports.OperationService) *InitiateDownloadUseCase {
	return &InitiateDownloadUseCase{
		coreService:      coreService,
		providerService:  providerAPI,
		operationService: operationService,
	}
}

func (uc *InitiateDownloadUseCase) Execute(ctx context.Context, song string, providerType string) (*model.OperationInitResult, error) {
	mediaDetails, err := uc.providerService.GetMediaDetails(ctx, song, providerType)
	if err != nil {
		return nil, errorsApp.ErrGetMediaDetailsFailed.WithMessage(fmt.Sprintf("error al obtener detalles del media: %v", err))
	}

	operationResult, err := uc.operationService.StartOperation(ctx, mediaDetails.ID)
	if err != nil {
		if errors.Is(err, errorsApp.ErrDuplicateRecord) {
			return nil, err
		}

		return nil, errorsApp.ErrStartOperationFailed.WithMessage(fmt.Sprintf("error al iniciar la operación: %v", err))
	}

	go func() {
		backgroundCtx := context.Background()
		err := uc.coreService.ProcessMedia(backgroundCtx, mediaDetails)
		if err != nil {
			fmt.Printf("Error en el procesamiento: %v", err)
		}
	}()

	return operationResult, nil
}
