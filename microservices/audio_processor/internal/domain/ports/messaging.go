package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
)

type (
	// MessageQueue es la interfaz que debe implementar cualquier servicio de mensajeria
	MessageQueue interface {
		SendMessage(ctx context.Context, message *model.MediaProcessingMessage) error
		ReceiveMessage(ctx context.Context) ([]model.MediaProcessingMessage, error)
		DeleteMessage(ctx context.Context, receiptHandle string) error
	}
)
