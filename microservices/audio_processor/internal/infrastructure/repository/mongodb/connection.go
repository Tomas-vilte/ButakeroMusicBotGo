package mongodb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
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
	fmt.Printf("URL DE MONDONGO: %s", uri)
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
	fmt.Println("Conexion exitosa")
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
	hostList := strings.Join(cfg.Database.Mongo.Host, ",")
	return fmt.Sprintf("mongodb://%s:%s@%s/?replicaSet=rs0",
		cfg.Database.Mongo.User,
		cfg.Database.Mongo.Password,
		hostList)
}
