package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

type MessageConsumer interface {
	ConsumeMessages(ctx context.Context, offset int64) error
	GetMessagesChannel() <-chan *entity.StatusMessage
}
