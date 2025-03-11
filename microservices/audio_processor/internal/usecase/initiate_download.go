package usecase

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
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
		return &model.OperationInitResult{}, fmt.Errorf("error al buscar el ID de la cancion: %w", err)
	}

	operationResult, err := uc.operationService.StartOperation(ctx, mediaDetails.ID)
	if err != nil {
		return nil, fmt.Errorf("error al iniciar la operación: %w", err)
	}

	go func() {
		backgroundCtx := context.Background()
		err := uc.coreService.ProcessMedia(backgroundCtx, operationResult.ID, mediaDetails)
		if err != nil {
			fmt.Printf("Error en el procesamiento: %v", err)
		}
	}()

	return operationResult, nil
}
