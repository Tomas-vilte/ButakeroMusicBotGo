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
	repo   ports.OperationRepository
	logger logger.Logger
}

func NewOperationService(repo ports.OperationRepository, logger logger.Logger) *OperationService {
	return &OperationService{
		repo:   repo,
		logger: logger,
	}
}

func (s *OperationService) StartOperation(ctx context.Context, songID string) (*model.OperationInitResult, error) {
	operation := &model.OperationResult{
		ID:     uuid.New().String(),
		SK:     songID,
		Status: statusInitiating,
	}

	if err := s.repo.SaveOperationsResult(ctx, operation); err != nil {
		s.logger.Error("Error al iniciar operación",
			zap.String("songID", songID),
			zap.Error(err))
		return nil, errors.ErrOperationInitFailed.Wrap(err)
	}

	return &model.OperationInitResult{
		ID:        operation.ID,
		SongID:    operation.SK,
		Status:    operation.Status,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}
