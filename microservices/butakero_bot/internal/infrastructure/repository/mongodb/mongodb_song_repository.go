package mongodb

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

// Options contiene las opciones de configuración para el repositorio
type Options struct {
	Collection     *mongo.Collection
	Database       string
	CollectionName string
	Logger         logging.Logger
}

// MongoSongRepository implementa la interfaz ports.SongRepository
type MongoSongRepository struct {
	opts Options
}

// NewMongoDBSongRepository crea una nueva instancia del repositorio
func NewMongoDBSongRepository(opts Options) (*MongoSongRepository, error) {
	if opts.Collection == nil {
		return nil, errors.New("mongo collection es requerido")
	}

	return &MongoSongRepository{
		opts: opts,
	}, nil
}

// GetSongByID obtiene una canción por su ID
func (r *MongoSongRepository) GetSongByID(ctx context.Context, id string) (*entity.Song, error) {
	var song entity.Song

	filter := bson.M{"_id": id}
	err := r.opts.Collection.FindOne(ctx, filter).Decode(&song)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.opts.Logger.Info("Cancion no encontrada", zap.String("id", id))
			return nil, nil
		}
		r.opts.Logger.Error("Error al obtener la cancion", zap.Error(err))
		return nil, err
	}

	r.opts.Logger.Info("Cancion obtenida exitosamente", zap.String("id", id))
	return &song, nil
}
