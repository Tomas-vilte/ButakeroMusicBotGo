package service

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"time"
)

type OperationManager struct {
	operationStore ports.OperationRepository
	logger         logger.Logger
	config         *config.Config
}

func NewOperationManager(o ports.OperationRepository, l logger.Logger, c *config.Config) *OperationManager {
	return &OperationManager{
		operationStore: o,
		logger:         l,
		config:         c,
	}
}

func (om *OperationManager) HandleOperationSuccess(ctx context.Context, operationID string, metadata *model.Metadata, fileData *model.FileData) error {
	result := &model.OperationResult{
		ID:             operationID,
		SK:             metadata.VideoID,
		Status:         statusSuccess,
		Message:        "Procesamiento completado exitosamente",
		Metadata:       metadata,
		FileData:       fileData,
		ProcessingDate: time.Now().Format(time.RFC3339),
		Success:        true,
		Attempts:       1,
		Failures:       0,
	}

	return om.operationStore.UpdateOperationResult(ctx, operationID, result)
}
