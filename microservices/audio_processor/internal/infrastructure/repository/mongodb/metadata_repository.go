package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var (
	// ErrMetadataNotFound se retorna cuando no se encuentran los metadatos
	ErrMetadataNotFound = errors.New("metadatos no encontrados en MongoDB")
	// ErrInvalidMetadata se retorna cuando los metadatos son inválidos
	ErrInvalidMetadata = errors.New("metadatos invalidos")
	// ErrDuplicateKey se retorna cuando se intenta insertar un documento con un ID que ya existe
	ErrDuplicateKey = errors.New("el ID ya existe en la base de datos")
)

type (
	// MongoMetadataRepository implementa la interfaz repository.MetadataRepository usando MongoDB
	MongoMetadataRepository struct {
		collection *mongo.Collection
		log        logger.Logger
	}

	// MongoMetadataOptions contiene las opciones para crear un nuevo MongoMetadataRepository
	MongoMetadataOptions struct {
		Log        logger.Logger
		Collection *mongo.Collection
	}
)

// NewMongoMetadataRepository crea una nueva instancia de MongoMetadataRepository
func NewMongoMetadataRepository(opts MongoMetadataOptions) (*MongoMetadataRepository, error) {
	if opts.Log == nil {
		return nil, fmt.Errorf("logger necesario")
	}

	if opts.Collection == nil {
		return nil, fmt.Errorf("collection necesario")
	}

	return &MongoMetadataRepository{
		collection: opts.Collection,
		log:        opts.Log,
	}, nil
}

// SaveMetadata guarda los metadatos en MongoDB
func (m *MongoMetadataRepository) SaveMetadata(ctx context.Context, metadata *model.Metadata) error {
	log := m.log.With(
		zap.String("component", "MongoMetadataRepository"),
		zap.String("method", "SaveMetadata"),
		zap.String("metadata_id", metadata.ID),
	)
	if err := validateMetadata(metadata); err != nil {
		log.Error("Metadatos inválidos", zap.Error(err))
		return fmt.Errorf("%w: %v", ErrInvalidMetadata, err)
	}

	if metadata.ID == "" {
		metadata.ID = uuid.New().String()
		log.Info("Generando nuevo ID para los metadatos", zap.String("new_id", metadata.ID))
	}

	doc := createMetadataDocument(metadata)
	log.Debug("Intentando guardar metadatos", zap.String("id", metadata.ID))

	_, err := m.collection.InsertOne(ctx, doc)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Error("ID duplicado al guardar metadatos", zap.Error(err))
			return ErrDuplicateKey
		}

		log.Error("Error al guardar metadatos", zap.Error(err))
		return fmt.Errorf("error al guardar metadatos: %w", err)
	}

	log.Info("Metadatos guardados exitosamente", zap.String("id", metadata.ID))
	return nil
}

// GetMetadata recupera los metadatos por ID
func (m *MongoMetadataRepository) GetMetadata(ctx context.Context, id string) (*model.Metadata, error) {
	log := m.log.With(
		zap.String("component", "MongoMetadataRepository"),
		zap.String("method", "GetMetadata"),
		zap.String("metadata_id", id),
	)
	if id == "" {
		log.Error("ID inválido", zap.String("id", id))
		return nil, errors.New("ID no puede estar vacio")
	}

	log.Debug("Buscando metadatos", zap.String("id", id))

	var metadata model.Metadata
	filter := bson.D{
		{"_id", id},
	}

	err := m.collection.FindOne(ctx, filter).Decode(&metadata)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Debug("Metadatos no encontrados", zap.String("id", id))
			return nil, ErrMetadataNotFound
		}
		log.Error("Error al recuperar metadatos", zap.Error(err))
		return nil, fmt.Errorf("error al recuperar metadatos: %w", err)
	}

	log.Debug("Metadatos recuperados exitosamente", zap.String("id", id))
	return &metadata, nil
}

// DeleteMetadata elimina los metadatos por ID
func (m *MongoMetadataRepository) DeleteMetadata(ctx context.Context, id string) error {
	log := m.log.With(
		zap.String("component", "MongoMetadataRepository"),
		zap.String("method", "DeleteMetadata"),
		zap.String("metadata_id", id),
	)
	if id == "" {
		log.Error("ID inválido", zap.String("id", id))
		return errors.New("ID no puede estar vacio")
	}

	log.Debug("Intentando eliminar metadatos", zap.String("id", id))

	filter := bson.M{"_id": id}
	result, err := m.collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error("Error al eliminar metadatos", zap.Error(err))
		return fmt.Errorf("error al eliminar metadatos: %w", err)
	}

	if result.DeletedCount == 0 {
		log.Debug("No se encontraron metadatos para eliminar", zap.String("id", id))
		return ErrMetadataNotFound
	}

	log.Info("Metadatos eliminados exitosamente", zap.String("id", id))
	return nil
}

// createMetadataDocument crea un documento BSON a partir de los metadatos
func createMetadataDocument(metadata *model.Metadata) bson.M {
	now := time.Now()
	return bson.M{
		"_id":           metadata.ID,
		"title":         metadata.Title,
		"url":           metadata.URL,
		"thumbnail_url": metadata.ThumbnailURL,
		"platform":      metadata.Platform,
		"duration":      metadata.Duration,
		"createdAt":     now,
		"updatedAt":     now,
	}
}

// validateMetadata valida que los campos requeridos estén presentes
func validateMetadata(metadata *model.Metadata) error {
	if metadata.Title == "" {
		return errors.New("título es requerido")
	}
	if metadata.URL == "" {
		return errors.New("URL de YouTube es requerida")
	}
	if metadata.Platform == "" {
		return errors.New("plataforma es requerida")
	}
	return nil
}
