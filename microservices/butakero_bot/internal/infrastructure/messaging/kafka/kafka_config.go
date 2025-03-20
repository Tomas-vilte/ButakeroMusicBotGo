package kafka

import "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared"

type (
	KafkaConfig struct {
		Brokers []string
		Topic   string
		TLS     shared.TLSConfig
	}
)
