package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"time"
)

type DownloadProcessor struct {
	consumer         ports.MessageConsumer
	mediaService     ports.MediaService
	videoService     ports.VideoService
	coreService      ports.CoreService
	operationService ports.OperationService
	logger           logger.Logger
	maxRetries       int
	retryDelay       time.Duration
}

func NewDownloadProcessor(
	consumer ports.MessageConsumer,
	mediaService ports.MediaService,
	videoService ports.VideoService,
	coreService ports.CoreService,
	operationService ports.OperationService,
	logger logger.Logger,
) *DownloadProcessor {
	return &DownloadProcessor{
		consumer:         consumer,
		mediaService:     mediaService,
		videoService:     videoService,
		coreService:      coreService,
		operationService: operationService,
		logger:           logger,
		maxRetries:       3,
		retryDelay:       2 * time.Second,
	}
}

func (p *DownloadProcessor) Run(ctx context.Context) error {
	requestChan, err := p.consumer.GetRequestsChannel(ctx)
	if err != nil {
		return fmt.Errorf("failed to get requests channel: %w", err)
	}

	p.logger.Info("Starting download processor")

	for {
		select {
		case req := <-requestChan:
			if err := p.processRequest(ctx, req); err != nil {
				p.logger.Error("Failed to process request",
					zap.String("interaction_id", req.InteractionID),
					zap.Error(err))
			}
		case <-ctx.Done():
			p.logger.Info("Shutdown signal received, stopping processor")
			return nil
		}
	}
}

func (p *DownloadProcessor) processRequest(ctx context.Context, req *model.MediaRequest) error {
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	log := p.logger.With(
		zap.String("interaction_id", req.InteractionID),
		zap.String("user_id", req.UserID),
	)

	mediaDetails, err := p.videoService.GetMediaDetails(reqCtx, req.Song, req.ProviderType)
	if err != nil {
		log.Error("Failed to get media details", zap.Error(err))
		return err
	}

	_, err = p.operationService.StartOperation(ctx, mediaDetails)
	if err != nil {
		return err
	}

	if err := p.coreService.ProcessMedia(reqCtx, mediaDetails); err != nil {
		log.Error("Media processing failed", zap.Error(err))
		return err
	}

	log.Info("Media processed successfully")
	return nil
}
