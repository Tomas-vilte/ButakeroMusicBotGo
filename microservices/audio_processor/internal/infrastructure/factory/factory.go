package factory

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/port"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue/kafka"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue/sqs"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/repository/dynamodb"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/repository/mongodb"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/storage/cloud"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/storage/local"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
)

type (
	AWSFactory struct{}

	LocalFactory struct{}
)

func (f *AWSFactory) CreateStorage(cfg *config.Config) (port.Storage, error) {
	return cloud.NewS3Storage(cfg)
}

func (f *AWSFactory) CreateQueue(cfg *config.Config, log logger.Logger) (port.MessageQueue, error) {
	return sqs.NewSQSService(cfg, log)
}

func (f *AWSFactory) CreateMetadataRepository(cfg *config.Config, log logger.Logger) (port.MetadataRepository, error) {
	return dynamodb.NewMetadataStore(cfg)
}

func (f *AWSFactory) CreateOperationRepository(cfg *config.Config, log logger.Logger) (port.OperationRepository, error) {
	return dynamodb.NewOperationStore(cfg)
}

func (f *LocalFactory) CreateStorage(cfg *config.Config) (port.Storage, error) {
	return local.NewLocalStorage(cfg)
}

func (f *LocalFactory) CreateQueue(cfg *config.Config, log logger.Logger) (port.MessageQueue, error) {
	return kafka.NewKafkaService(cfg, log)
}

func (f *LocalFactory) CreateMetadataRepository(cfg *config.Config, log logger.Logger) (port.MetadataRepository, error) {
	conn, err := mongodb.NewMongoDB(mongodb.MongoOptions{
		Config: cfg,
		Log:    log,
	})
	if err != nil {
		return nil, err
	}

	return mongodb.NewMongoMetadataRepository(mongodb.MongoMetadataOptions{
		Log:        log,
		Collection: conn.GetCollection(cfg.Database.Mongo.Collections.Songs),
	})
}

func (f *LocalFactory) CreateOperationRepository(cfg *config.Config, log logger.Logger) (port.OperationRepository, error) {
	conn, err := mongodb.NewMongoDB(mongodb.MongoOptions{
		Config: cfg,
		Log:    log,
	})
	if err != nil {
		return nil, err
	}

	return mongodb.NewOperationRepository(mongodb.OperationOptions{
		Log:        log,
		Collection: conn.GetCollection(cfg.Database.Mongo.Collections.Operations),
	})
}
