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
	mediaService ports.MediaService
	videoService ports.VideoService
	coreService  ports.CoreService
	logger       logger.Logger
}

func NewMediaProcessor(
	mediaService ports.MediaService,
	videoService ports.VideoService,
	coreService ports.CoreService,
	logger logger.Logger,
) *MediaProcessor {
	return &MediaProcessor{
		mediaService: mediaService,
		videoService: videoService,
		coreService:  coreService,
		logger:       logger,
	}
}

func (p *MediaProcessor) ProcessRequest(ctx context.Context, req *model.MediaRequest) error {
	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	log := p.logger.With(
		zap.String("request_id", req.RequestID),
		zap.String("user_id", req.UserID),
	)

	mediaDetails, err := p.videoService.GetMediaDetails(reqCtx, req.Song, req.ProviderType)
	if err != nil {
		log.Error("Error al obtener detalles del media", zap.Error(err))
		return err
	}

	media := &model.Media{
		VideoID:    mediaDetails.ID,
		Status:     "starting",
		TitleLower: "",
		Metadata: &model.PlatformMetadata{
			Title:        mediaDetails.Title,
			DurationMs:   0,
			URL:          "",
			ThumbnailURL: "",
			Platform:     "",
		},
		FileData:       &model.FileData{},
		ProcessingDate: time.Now(),
		Success:        false,
		Attempts:       0,
		Failures:       0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := p.mediaService.CreateMedia(ctx, media); err != nil {
		log.Error("Error al crear registro inicial", zap.Error(err))
		return err
	}

	if err := p.coreService.ProcessMedia(reqCtx, mediaDetails, req.UserID, req.RequestID); err != nil {
		log.Error("Error al procesar media", zap.Error(err))
		return err
	}

	log.Info("Media procesado exitosamente")
	return nil
}
