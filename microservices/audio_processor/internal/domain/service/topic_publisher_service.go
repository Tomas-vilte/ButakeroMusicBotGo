package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
)

type topicPublisherService struct {
	messageQueue ports.MessageQueue
	logger       logger.Logger
}

func NewMediaProcessingPublisherService(messageQueue ports.MessageQueue, logger logger.Logger) ports.TopicPublisherService {
	return &topicPublisherService{
		messageQueue: messageQueue,
		logger:       logger,
	}
}

func (s *topicPublisherService) PublishMediaProcessed(ctx context.Context, message *model.MediaProcessingMessage) error {
	log := s.logger.With(
		zap.String("component", "TopicPublisherService"),
		zap.String("method", "PublishMediaProcessed"),
		zap.String("video_id", message.VideoID),
	)

	if err := s.messageQueue.SendMessage(ctx, message); err != nil {
		log.Error("Error al publicar el mensaje", zap.Error(err))
		return fmt.Errorf("error al publicar el mensaje: %w", err)
	}

	log.Info("Mensaje publicado exitosamente", zap.String("video_id", message.VideoID))
	return nil
}
