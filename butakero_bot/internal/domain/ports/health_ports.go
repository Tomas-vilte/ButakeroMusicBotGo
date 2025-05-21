package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

type (
	// HealthChecker es una interfaz que define un método para verificar la salud de un servicio
	HealthChecker[T any] interface {
		// Check devuelve un valor de tipo T y un error
		Check(ctx context.Context) (T, error)
	}

	// DiscordHealthChecker es una interfaz que extiende la interfaz HealthChecker
	DiscordHealthChecker interface {
		// HealthChecker es un método que devuelve un valor de tipo entity.DiscordHealth y un error
		HealthChecker[entity.DiscordHealth]
	}

	// ServiceBHealthChecker es una interfaz que extiende la interfaz HealthChecker
	ServiceBHealthChecker interface {
		// HealthChecker es un método que devuelve un valor de tipo entity.ServiceBHealth y un error
		HealthChecker[entity.ServiceBHealth]
	}
)
