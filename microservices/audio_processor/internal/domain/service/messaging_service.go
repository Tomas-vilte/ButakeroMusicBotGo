package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"time"
)

type MessagingService struct {
	messaging ports.MessageQueue
	logger    logger.Logger
}

func NewMessagingService(m ports.MessageQueue, l logger.Logger) *MessagingService {
	return &MessagingService{
		messaging: m,
		logger:    l,
	}
}

func (ms *MessagingService) SendProcessingMessage(ctx context.Context, operationID string, status string, metadata *model.Metadata, attempts int) error {
	message := model.Message{
		ID:      operationID,
		Content: fmt.Sprintf("Procesamiento %s", status),
		Status: model.Status{
			ID:             operationID,
			SK:             metadata.VideoID,
			Status:         status,
			Metadata:       metadata,
			ProcessingDate: time.Now().UTC(),
			Attempts:       attempts,
		},
	}

	if err := ms.messaging.SendMessage(ctx, message); err != nil {
		return errors.ErrMessageSendFailed.Wrap(err)
	}
	return nil
}
