package service

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

type OperationService struct {
	repo   ports.MediaRepository
	logger logger.Logger
}

func NewOperationService(repo ports.MediaRepository, logger logger.Logger) *OperationService {
	return &OperationService{
		repo:   repo,
		logger: logger,
	}
}

func (s *OperationService) StartOperation(ctx context.Context, videoID string) (*model.OperationInitResult, error) {
	media := &model.Media{
		ID:      uuid.New().String(),
		VideoID: videoID,
		Status:  "starting",
		Metadata: &model.PlatformMetadata{
			Title:        "",
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

	if err := s.repo.SaveMedia(ctx, media); err != nil {
		s.logger.Error("Error al iniciar operación",
			zap.String("songID", videoID),
			zap.Error(err))
		return nil, errors.ErrOperationInitFailed.Wrap(err)
	}

	return &model.OperationInitResult{
		ID:      media.ID,
		VideoID: media.VideoID,
		Status:  media.Status,
	}, nil
}
