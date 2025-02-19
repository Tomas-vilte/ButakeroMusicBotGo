package mongodb

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

// GetSongByVideoID obtiene una canción por su ID
func (r *MongoSongRepository) GetSongByVideoID(ctx context.Context, videoID string) (*entity.Song, error) {
	var song entity.Song

	filter := bson.M{"video_id": videoID}
	err := r.opts.Collection.FindOne(ctx, filter).Decode(&song)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			r.opts.Logger.Info("Cancion no encontrada", zap.String("id", videoID))
			return nil, nil
		}
		r.opts.Logger.Error("Error al obtener la cancion", zap.Error(err))
		return nil, err
	}

	r.opts.Logger.Info("Cancion obtenida exitosamente", zap.String("id", videoID))
	return &song, nil
}

func (r *MongoSongRepository) SearchSongsByTitle(ctx context.Context, title string) ([]*entity.Song, error) {
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{"title", "text"}},
		Options: options.Index().SetName("title_text"),
	}

	_, err := r.opts.Collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		r.opts.Logger.Warn("Error al crear el indice", zap.Error(err))
	}

	filter := bson.M{
		"$text": bson.M{
			"$search": title,
		},
	}

	findOptions := options.Find().
		SetSort(bson.D{{Key: "score", Value: bson.M{"$meta": "textScore"}}}).
		SetLimit(10)

	cursor, err := r.opts.Collection.Find(ctx, filter, findOptions)
	if err != nil {
		r.opts.Logger.Error("Error al buscar canciones", zap.Error(err))
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			r.opts.Logger.Error("Error al cerrar el cursor", zap.Error(err))
		}
	}()

	var songs []*entity.Song
	if err = cursor.All(ctx, &songs); err != nil {
		r.opts.Logger.Error("Error al decodificar canciones", zap.Error(err))
		return nil, err
	}

	return songs, nil
}
