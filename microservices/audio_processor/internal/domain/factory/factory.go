package factory

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/port"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
)

type EnvironmentFactory interface {
	CreateStorage(cfg *config.Config) (port.Storage, error)
	CreateQueue(cfg *config.Config, log logger.Logger) (port.MessageQueue, error)
	CreateMetadataRepository(cfg *config.Config, log logger.Logger) (port.MetadataRepository, error)
	CreateOperationRepository(cfg *config.Config, log logger.Logger) (port.OperationRepository, error)
}
