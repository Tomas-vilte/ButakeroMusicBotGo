package mongodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	errorsApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"time"
)

type (
	// MediaRepository implementa la interfaz MediaRepository para MongoDB.
	MediaRepository struct {
		collection *mongo.Collection
		log        logger.Logger
	}

	// MediaRepositoryOptions contiene las opciones para crear un nuevo MediaRepository.
	MediaRepositoryOptions struct {
		Collection *mongo.Collection
		Log        logger.Logger
	}
)

// NewMediaRepository crea una nueva instancia de MediaRepository.
func NewMediaRepository(opts MediaRepositoryOptions) (*MediaRepository, error) {
	if opts.Collection == nil {
		return nil, errorsApp.ErrInvalidInput.WithMessage("collection no puede ser nil")
	}
	if opts.Log == nil {
		return nil, errorsApp.ErrInvalidInput.WithMessage("logger no puede ser nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "title_lower", Value: "text"}},
		Options: options.Index().SetName("title_text"),
	}

	_, err := opts.Collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		opts.Log.Warn("Error al crear el índice", zap.Error(err))
	}

	return &MediaRepository{
		collection: opts.Collection,
		log:        opts.Log,
	}, nil
}

// SaveMedia guarda un registro de procesamiento multimedia en MongoDB.
func (r *MediaRepository) SaveMedia(ctx context.Context, media *model.Media) error {
	log := r.log.With(
		zap.String("component", "MediaRepository"),
		zap.String("method", "SaveMedia"),
	)

	_, err := r.collection.InsertOne(ctx, media)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn("ID duplicado al guardar el registro de media, intentando actualizar", zap.Error(err))
			return errorsApp.ErrDuplicateRecord.WithMessage(fmt.Sprintf("El video con ID '%s' ya está registrado.", media.VideoID), media.VideoID)
		}
		log.Error("Error al guardar el registro de media", zap.Error(err))
		return errorsApp.ErrCodeSaveMediaFailed.WithMessage(fmt.Sprintf("error al guardar el registro de media: %v", err))
	}

	log.Info("Registro de media guardado exitosamente")
	return nil
}

func (r *MediaRepository) GetMediaByTitle(ctx context.Context, title string) ([]*model.Media, error) {
	log := r.log.With(
		zap.String("method", "SearchSongsByTitle"),
		zap.String("title", title),
	)

	result := make([]*model.Media, 0)

	filter := bson.M{
		"title_lower": bson.M{
			"$regex":   fmt.Sprintf(".*%s.*", title),
			"$options": "i",
		},
	}

	findOptions := options.Find().
		SetLimit(10)

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		log.Error("Error al buscar canciones", zap.Error(err))
		return result, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Error("Error al cerrar el cursor", zap.Error(err))
		}
	}()

	if err = cursor.All(ctx, &result); err != nil {
		log.Error("Error al decodificar canciones", zap.Error(err))
		return result, err
	}

	log.Info("Búsqueda de canciones completada", zap.Int("count", len(result)))
	return result, nil
}

// GetMediaByID obtiene un registro de procesamiento multimedia por su ID y video_id.
func (r *MediaRepository) GetMediaByID(ctx context.Context, videoID string) (*model.Media, error) {
	log := r.log.With(
		zap.String("component", "MediaRepository"),
		zap.String("method", "GetMediaByID"),
		zap.String("video_id", videoID),
	)

	filter := bson.D{
		{Key: "_id", Value: videoID},
	}

	var media model.Media
	err := r.collection.FindOne(ctx, filter).Decode(&media)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Warn("Registro de media no encontrado")
			return nil, errorsApp.ErrCodeMediaNotFound.WithMessage(
				"Media no encontrado",
				videoID,
			)
		}
		log.Error("Error al obtener el registro de media", zap.Error(err))
		return nil, errorsApp.ErrGetMediaDetailsFailed.WithMessage(fmt.Sprintf("error al obtener el registro de media: %v", err))
	}

	log.Info("Registro de media recuperado exitosamente")
	return &media, nil
}

// DeleteMedia elimina un registro de procesamiento multimedia por su ID y video_id.
func (r *MediaRepository) DeleteMedia(ctx context.Context, videoID string) error {
	log := r.log.With(
		zap.String("component", "MediaRepository"),
		zap.String("method", "DeleteMedia"),
		zap.String("video_id", videoID),
	)

	filter := bson.D{
		{Key: "_id", Value: videoID},
	}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error("Error al eliminar el registro de media", zap.Error(err))
		return errorsApp.ErrCodeDeleteMediaFailed.WithMessage(fmt.Sprintf("error al eliminar el registro de media: %v", err))
	}

	if result.DeletedCount == 0 {
		log.Warn("Registro de media no encontrado para eliminar")
		return errorsApp.ErrCodeMediaNotFound
	}

	log.Info("Registro de media eliminado exitosamente")
	return nil
}

// UpdateMedia actualiza un registro de procesamiento multimedia de manera más eficiente.
func (r *MediaRepository) UpdateMedia(ctx context.Context, videoID string, media *model.Media) error {
	log := r.log.With(
		zap.String("component", "MediaRepository"),
		zap.String("method", "UpdateMedia"),
		zap.String("video_id", videoID),
	)

	filter := bson.D{
		{Key: "_id", Value: videoID},
	}

	opts := options.Update().SetUpsert(false)

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "title_lower", Value: media.TitleLower},
			{Key: "status", Value: media.Status},
			{Key: "message", Value: media.Message},
			{Key: "metadata", Value: media.Metadata},
			{Key: "file_data", Value: media.FileData},
			{Key: "processing_date", Value: media.ProcessingDate},
			{Key: "success", Value: media.Success},
			{Key: "attempts", Value: media.Attempts},
			{Key: "failures", Value: media.Failures},
			{Key: "updated_at", Value: media.UpdatedAt},
			{Key: "play_count", Value: media.PlayCount},
		}},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Error("Error al actualizar el registro de media", zap.Error(err))
		return errorsApp.ErrUpdateMediaFailed.WithMessage(fmt.Sprintf("error al actualizar el registro de media: %v", err))
	}

	if result.MatchedCount == 0 && result.UpsertedCount == 0 {
		log.Warn("Registro de media no encontrado para actualizar")
		return errorsApp.ErrCodeMediaNotFound
	}

	log.Info("Registro de media actualizado exitosamente")
	return nil
}
