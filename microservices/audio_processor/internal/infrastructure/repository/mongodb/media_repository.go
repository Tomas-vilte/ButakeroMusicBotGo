package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	errorsApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

var (
	ErrMediaNotFound   = errors.New("registro de media no encontrado")
	ErrInvalidVideoID  = errors.New("video_id inválido")
	ErrInvalidMetadata = errors.New("metadatos inválidos")
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
		return nil, errors.New("collection no puede ser nil")
	}
	if opts.Log == nil {
		return nil, errors.New("logger no puede ser nil")
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

	if media.VideoID == "" {
		log.Error("video_id no puede estar vacío")
		return ErrInvalidVideoID
	}

	now := time.Now()
	media.CreatedAt = now
	media.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, media)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn("ID duplicado al guardar el registro de media, intentando actualizar", zap.Error(err))
			return errorsApp.ErrDuplicateRecord.WithMessage(fmt.Sprintf("El video con ID '%s' ya está registrado.", media.VideoID))
		}
		log.Error("Error al guardar el registro de media", zap.Error(err))
		return fmt.Errorf("error al guardar el registro de media: %w", err)
	}

	log.Info("Registro de media guardado exitosamente")
	return nil
}

// GetMedia obtiene un registro de procesamiento multimedia por su ID y video_id.
func (r *MediaRepository) GetMedia(ctx context.Context, videoID string) (*model.Media, error) {
	log := r.log.With(
		zap.String("component", "MediaRepository"),
		zap.String("method", "GetMedia"),
		zap.String("video_id", videoID),
	)

	if videoID == "" {
		log.Error("video_id no puede estar vacío")
		return nil, ErrInvalidVideoID
	}

	filter := bson.D{
		{Key: "_id", Value: videoID},
	}

	var media model.Media
	err := r.collection.FindOne(ctx, filter).Decode(&media)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Warn("Registro de media no encontrado")
			return nil, ErrMediaNotFound
		}
		log.Error("Error al obtener el registro de media", zap.Error(err))
		return nil, fmt.Errorf("error al obtener el registro de media: %w", err)
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

	if videoID == "" {
		log.Error("video_id no puede estar vacío")
		return ErrInvalidVideoID
	}

	filter := bson.D{
		{Key: "_id", Value: videoID},
	}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error("Error al eliminar el registro de media", zap.Error(err))
		return fmt.Errorf("error al eliminar el registro de media: %w", err)
	}

	if result.DeletedCount == 0 {
		log.Warn("Registro de media no encontrado para eliminar")
		return ErrMediaNotFound
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

	if videoID == "" {
		log.Error("video_id no puede estar vacío")
		return ErrInvalidVideoID
	}

	if media.Metadata == nil || media.Metadata.Title == "" || media.Metadata.Platform == "" {
		log.Error("Metadatos inválidos", zap.Any("metadata", media.Metadata))
		return ErrInvalidMetadata
	}

	media.UpdatedAt = time.Now()

	media.VideoID = videoID

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
		return fmt.Errorf("error al actualizar el registro de media: %w", err)
	}

	if result.MatchedCount == 0 && result.UpsertedCount == 0 {
		log.Warn("Registro de media no encontrado para actualizar")
		return ErrMediaNotFound
	}

	log.Info("Registro de media actualizado exitosamente")
	return nil
}
