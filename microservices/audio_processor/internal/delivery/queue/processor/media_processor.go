package processor

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"time"
)

type MediaProcessor struct {
	mediaService     ports.MediaService
	videoService     ports.VideoService
	coreService      ports.CoreService
	operationService ports.OperationService
	logger           logger.Logger
}

func NewMediaProcessor(
	mediaService ports.MediaService,
	videoService ports.VideoService,
	coreService ports.CoreService,
	operationService ports.OperationService,
	logger logger.Logger,
) *MediaProcessor {
	return &MediaProcessor{
		mediaService:     mediaService,
		videoService:     videoService,
		coreService:      coreService,
		operationService: operationService,
		logger:           logger,
	}
}

func (p *MediaProcessor) ProcessRequest(ctx context.Context, req *model.MediaRequest) error {
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	log := p.logger.With(
		zap.String("interaction_id", req.InteractionID),
		zap.String("user_id", req.UserID),
	)

	mediaDetails, err := p.videoService.GetMediaDetails(reqCtx, req.Song, req.ProviderType)
	if err != nil {
		log.Error("Error al obtener detalles del media", zap.Error(err))
		return err
	}

	_, err = p.operationService.StartOperation(ctx, mediaDetails)
	if err != nil {
		return err
	}

	if err := p.coreService.ProcessMedia(reqCtx, mediaDetails); err != nil {
		log.Error("Error al procesar media", zap.Error(err))
		return err
	}

	log.Info("Media procesado exitosamente")
	return nil
}
