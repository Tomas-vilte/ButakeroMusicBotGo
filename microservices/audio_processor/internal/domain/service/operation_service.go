package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	errorsApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

type operationService struct {
	repo   ports.MediaRepository
	logger logger.Logger
}

func NewOperationService(repo ports.MediaRepository, logger logger.Logger) ports.OperationService {
	return &operationService{
		repo:   repo,
		logger: logger,
	}
}

func (s *operationService) StartOperation(ctx context.Context, videoID string) (*model.OperationInitResult, error) {
	media := &model.Media{
		VideoID:    videoID,
		Status:     "starting",
		TitleLower: "",
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

		if errors.Is(err, errorsApp.ErrDuplicateRecord) {
			return nil, errorsApp.ErrDuplicateRecord.WithMessage(fmt.Sprintf("El video con ID '%s' ya esta registado", videoID))
		}
		return nil, errorsApp.ErrOperationInitFailed.Wrap(err)
	}

	return &model.OperationInitResult{
		VideoID: media.VideoID,
		Status:  media.Status,
	}, nil
}
