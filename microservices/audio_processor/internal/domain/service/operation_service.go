package service

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"time"
)

type operationService struct {
	mediaService ports.MediaService
	logger       logger.Logger
}

func NewOperationService(mediaService ports.MediaService, logger logger.Logger) ports.OperationService {
	return &operationService{
		mediaService: mediaService,
		logger:       logger,
	}
}

func (s *operationService) StartOperation(ctx context.Context, mediaDetails *model.MediaDetails) (*model.OperationInitResult, error) {
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
		FileData: &model.FileData{
			FilePath: "",
			FileSize: "",
			FileType: "",
		},
		ProcessingDate: time.Now(),
		Success:        false,
		Attempts:       0,
		Failures:       0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		PlayCount:      0,
	}

	if err := s.mediaService.CreateMedia(ctx, media); err != nil {
		s.logger.Error("Error al iniciar operación",
			zap.String("songID", media.VideoID),
			zap.Error(err))
		return nil, err
	}

	return &model.OperationInitResult{
		VideoID: media.VideoID,
		Status:  media.Status,
	}, nil
}
