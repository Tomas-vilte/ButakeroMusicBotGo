package mongodb

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/utils"
	"github.com/pkg/errors"
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

	var tlsConfig *tls.Config
	var err error

	if opts.Config.Database.Mongo.EnableTLS {
		tlsConfig, err = utils.NewTLSConfig(&utils.TLSConfig{
			CaFile:   opts.Config.Database.Mongo.CaFile,
			CertFile: opts.Config.Database.Mongo.CertFile,
			KeyFile:  opts.Config.Database.Mongo.KeyFile,
		})
		if err != nil {
			return nil, errors.Wrap(err, "Error configurando conexion de TLS de MongoDB")
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uri := utils.BuildMongoURI(opts.Config)
	clientOptions := options.Client().ApplyURI(uri)

	if opts.Config.Database.Mongo.EnableTLS {
		clientOptions.SetTLSConfig(tlsConfig)
	}

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		opts.Log.Error("Error al conectar a MongoDB", zap.Error(err))
		return nil, err
	}

	err = client.Ping(ctx, readpref.PrimaryPreferred())
	if err != nil {
		return nil, fmt.Errorf("error al hacer ping a MongoDB: %w", err)
	}

	opts.Log.Info("Conexion exitosa a MongoDB", zap.String("Database", opts.Config.Database.Mongo.Database),
		zap.Strings("Collections", []string{opts.Config.Database.Mongo.Collections.Songs}),
		zap.Strings("Hosts", opts.Config.Database.Mongo.Host),
		zap.String("ReplicaSet", opts.Config.Database.Mongo.ReplicaSetName),
		zap.Bool("TLS", opts.Config.Database.Mongo.EnableTLS))
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
