package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
)

type (
	MessageConsumer interface {
		GetRequestsChannel(ctx context.Context) (<-chan *model.MediaRequest, error)
		Close() error
	}

	MessageProducer interface {
		Publish(ctx context.Context, message *model.MediaProcessingMessage) error
		Close() error
	}
)
