package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

type HealthChecker[T any] interface {
	Check(ctx context.Context) (T, error)
}

type DiscordHealthChecker interface {
	HealthChecker[entity.DiscordHealth]
}

type ServiceBHealthChecker interface {
	HealthChecker[entity.ServiceBHealth]
}
