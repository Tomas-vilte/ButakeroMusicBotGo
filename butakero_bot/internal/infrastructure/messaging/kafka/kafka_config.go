package kafka

import "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared"

type (
	ConfigKafka struct {
		Brokers []string
		Topic   string
		TLS     shared.TLSConfig
		Offset  int64
	}
)
