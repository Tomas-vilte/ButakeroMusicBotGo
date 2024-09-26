package usecase

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/service"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
	"sync"
)

type InitiateDownloadUseCase struct {
	audioService service.AudioProcessor
	youtubeAPI   api.YouTubeService
}

func NewInitiateDownloadUseCase(audioService service.AudioProcessor, youtubeAPI api.YouTubeService) *InitiateDownloadUseCase {
	return &InitiateDownloadUseCase{
		audioService: audioService,
		youtubeAPI:   youtubeAPI,
	}
}

func (uc *InitiateDownloadUseCase) Execute(ctx context.Context, song string) (string, error) {
	var wg sync.WaitGroup
	videoID, err := uc.youtubeAPI.SearchVideoID(ctx, song)
	if err != nil {
		return "", fmt.Errorf("error al buscar el ID de la cancion: %w", err)
	}

	youtubeMetadata, err := uc.youtubeAPI.GetVideoDetails(ctx, videoID)
	if err != nil {
		return "", fmt.Errorf("error al obtener metadata de YouTube: %w", err)
	}

	operationID, err := uc.audioService.StartOperation(ctx, videoID)
	if err != nil {
		return "", fmt.Errorf("error al iniciar la operación: %w", err)
	}
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := uc.audioService.ProcessAudio(ctx, operationID, *youtubeMetadata)
		if err != nil {
			fmt.Printf("Error en el procesamiento: %v", err)
		}
	}()
	wg.Wait()
	return operationID, nil
}
