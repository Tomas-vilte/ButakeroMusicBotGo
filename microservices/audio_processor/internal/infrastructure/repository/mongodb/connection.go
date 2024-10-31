package mongodb

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
	"time"
)

type (
	MongoDB struct {
		client *mongo.Client
		config *config.Config
		log    logger.Logger
	}

	MongoOptions struct {
		Config *config.Config
		Log    logger.Logger
	}
)

func NewMongoDB(opts MongoOptions) (*MongoDB, error) {
	if opts.Log == nil {
		return nil, fmt.Errorf("logger necesario")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := buildMongoURI(opts.Config)
	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		opts.Log.Error("Error al conectar a MongoDB", zap.Error(err))
		return nil, err
	}

	err = client.Ping(ctx, readpref.PrimaryPreferred())
	if err != nil {
		return nil, fmt.Errorf("error al hacer ping a MongoDB: %w", err)
	}
	return &MongoDB{
		client: client,
		config: opts.Config,
		log:    opts.Log,
	}, nil
}

func (db *MongoDB) GetCollection(collectionName string) *mongo.Collection {
	return db.client.Database(db.config.Database.Mongo.Database).Collection(collectionName)
}

func (db *MongoDB) Close(ctx context.Context) error {
	return db.client.Disconnect(ctx)
}

func buildMongoURI(cfg *config.Config) string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s",
		cfg.Database.Mongo.User,
		cfg.Database.Mongo.Password,
		cfg.Database.Mongo.Host,
		cfg.Database.Mongo.Port)
}
