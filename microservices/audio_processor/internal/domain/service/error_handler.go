package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"time"
)

type ErrorHandler struct {
	opsRepo   ports.OperationRepository
	messaging ports.MessageQueue
	logger    logger.Logger
	config    *config.Config
}

func NewErrorHandler(opsRepo ports.OperationRepository, messaging ports.MessageQueue, logger logger.Logger, cfg *config.Config) *ErrorHandler {
	return &ErrorHandler{
		opsRepo:   opsRepo,
		messaging: messaging,
		logger:    logger,
		config:    cfg,
	}
}

func (eh *ErrorHandler) HandleProcessingError(
	ctx context.Context,
	operationID string,
	metadata *model.Metadata,
	errorType string,
	attempts int,
	originalErr error,
) error {
	result := &model.OperationResult{
		ID:             operationID,
		SK:             metadata.VideoID,
		Status:         statusFailed,
		Message:        fmt.Sprintf("[%s] %v", errorType, originalErr),
		Metadata:       metadata,
		ProcessingDate: time.Now().Format(time.RFC3339),
		Attempts:       attempts,
		Failures:       attempts,
	}

	if err := eh.opsRepo.UpdateOperationResult(ctx, operationID, result); err != nil {
		eh.logger.Error("Error al actualizar estado fallido", zap.Error(err))
	}

	message := model.Message{
		ID:      operationID,
		Content: "Error en procesamiento",
		Status: model.Status{
			ID:             operationID,
			SK:             metadata.VideoID,
			Status:         statusFailed,
			Message:        result.Message,
			Metadata:       metadata,
			ProcessingDate: time.Now().UTC(),
			Attempts:       attempts,
			Failures:       attempts,
		},
	}

	if err := eh.messaging.SendMessage(ctx, message); err != nil {
		eh.logger.Error("Error al enviar mensaje de fallo", zap.Error(err))
	}

	return originalErr
}
