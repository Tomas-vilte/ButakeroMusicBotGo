package mongodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var (
	ErrOperationNotFound = errors.New("operación no encontrada")
	ErrInvalidUUID       = errors.New("UUID inválido")
	ErrInvalidStatus     = errors.New("estado inválido")
)

// ValidStatus define los estados permitidos para una operación.
var ValidStatus = map[string]bool{
	"starting": true,
	"failed":   true,
	"success":  true,
}

type (
	// OperationRepository define el repositorio de operaciones con MongoDB.
	OperationRepository struct {
		collection *mongo.Collection
		log        logger.Logger
	}

	// OperationOptions agrupa las opciones necesarias para inicializar OperationRepository.
	OperationOptions struct {
		Collection *mongo.Collection
		Log        logger.Logger
	}
)

// isValidUUID verifica si una cadena es un UUID válido
func isValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// createSafeFilter crea un filtro BSON para las consultas, usando IDs válidos
func createSafeFilter(id, sk string) (bson.D, error) {
	// solo validar `id` como UUID
	if !isValidUUID(id) {
		return nil, ErrInvalidUUID
	}

	if sk == "" {
		return nil, errors.New("sk no puede estar vacio")
	}

	return bson.D{
		{Key: "_id", Value: id},
		{Key: "sk", Value: sk},
	}, nil
}

// NewOperationRepository crea un nuevo repositorio de operaciones con las opciones proporcionadas.
func NewOperationRepository(opts OperationOptions) (*OperationRepository, error) {
	if opts.Collection == nil {
		return nil, errors.New("collection no puede ser nil")
	}
	if opts.Log == nil {
		return nil, errors.New("logger no puede ser nil")
	}

	return &OperationRepository{
		collection: opts.Collection,
		log:        opts.Log,
	}, nil
}

// SaveOperationsResult guarda un resultado de operación en la colección MongoDB.
func (s *OperationRepository) SaveOperationsResult(ctx context.Context, result *model.OperationResult) error {
	log := s.log.With(
		zap.String("component", "OperationRepository"),
		zap.String("method", "SaveOperationsResult"),
		zap.String("operation_id", result.ID),
	)

	if result.ID == "" {
		result.ID = uuid.New().String()
		log.Info("Generando nuevo ID para la operación", zap.String("new_id", result.ID))

	}

	if !isValidUUID(result.ID) {
		log.Error("UUID inválido", zap.String("operation_id", result.ID))
		return ErrInvalidUUID
	}

	if _, err := s.collection.InsertOne(ctx, result); err != nil {
		log.Error("Error al guardar resultado de operación", zap.Error(err))
		return fmt.Errorf("error al guardar resultado de operación: %w", err)
	}
	log.Info("Operación guardada exitosamente", zap.String("id", result.ID))
	return nil
}

// GetOperationResult obtiene el resultado de una operación a partir de su ID y sk.
func (s *OperationRepository) GetOperationResult(ctx context.Context, id, sk string) (*model.OperationResult, error) {
	log := s.log.With(
		zap.String("component", "OperationRepository"),
		zap.String("method", "GetOperationResult"),
		zap.String("operation_id", id),
		zap.String("sk", sk),
	)
	if sk == "" {
		log.Error("SK no puede estar vacío")
		return nil, fmt.Errorf("sk no puede estar vacia")
	}

	filter, err := createSafeFilter(id, sk)
	if err != nil {
		log.Error("Error al crear el filtro", zap.Error(err))
		return nil, err
	}
	var result model.OperationResult
	if err := s.collection.FindOne(ctx, filter).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Warn("Operación no encontrada", zap.String("id", id), zap.String("sk", sk))
			return nil, ErrOperationNotFound
		}
		log.Error("Error al recuperar operación", zap.Error(err))
		return nil, fmt.Errorf("error al recuperar operación: %w", err)
	}

	log.Info("Operación recuperada exitosamente", zap.String("id", id))
	return &result, nil
}

// DeleteOperationResult elimina el resultado de una operación específica en MongoDB.
func (s *OperationRepository) DeleteOperationResult(ctx context.Context, id, songID string) error {
	log := s.log.With(
		zap.String("component", "OperationRepository"),
		zap.String("method", "DeleteOperationResult"),
		zap.String("operation_id", id),
		zap.String("song_id", songID),
	)
	if songID == "" {
		log.Error("SongID no puede estar vacío")
		return fmt.Errorf("songID no puede estar vacio")
	}
	filter, err := createSafeFilter(id, songID)
	if err != nil {
		log.Error("Error al crear el filtro", zap.Error(err))
		return err
	}

	result, err := s.collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error("Error al eliminar la operación", zap.Error(err))
		return fmt.Errorf("error al eliminar el resultado de operacion desde MongoDB: %w", err)
	}

	if result.DeletedCount == 0 {
		log.Warn("Operación no encontrada para eliminar", zap.String("id", id))
		return ErrOperationNotFound
	}

	log.Info("Operación eliminada exitosamente", zap.String("id", id))
	return nil
}

// UpdateOperationStatus actualiza el estado de una operación, si el estado es válido.
func (s *OperationRepository) UpdateOperationStatus(ctx context.Context, operationID string, sk string, status string) error {
	log := s.log.With(
		zap.String("component", "OperationRepository"),
		zap.String("method", "UpdateOperationStatus"),
		zap.String("operation_id", operationID),
		zap.String("sk", sk),
		zap.String("status", status),
	)

	if !ValidStatus[status] {
		log.Error("Estado inválido", zap.String("status", status))
		return ErrInvalidStatus
	}

	filter, err := createSafeFilter(operationID, sk)
	if err != nil {
		log.Error("Error al crear el filtro", zap.Error(err))
		return err
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "status", Value: status},
		}},
	}

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error("Error al actualizar el estado de la operación", zap.Error(err))
		return fmt.Errorf("error al actualizar el estado de la operacion en MongoDB: %w", err)
	}

	if result.MatchedCount == 0 {
		log.Warn("Operación no encontrada para actualizar", zap.String("id", operationID))
		return ErrOperationNotFound
	}

	log.Info("Estado de la operación actualizado exitosamente", zap.String("id", operationID))
	return nil
}

func (s *OperationRepository) UpdateOperationResult(ctx context.Context, operationID string, operationResult *model.OperationResult) error {
	log := s.log.With(
		zap.String("component", "OperationRepository"),
		zap.String("method", "UpdateOperationResult"),
		zap.String("operation_id", operationID),
		zap.String("sk", operationResult.SK),
	)

	if operationResult.SK == "" {
		log.Error("SK no puede estar vacío")
		return errors.New("sk no puede estar vacio")
	}

	filter, err := createSafeFilter(operationID, operationResult.SK)
	if err != nil {
		log.Error("Error al crear el filtro", zap.Error(err))
		return err
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "status", Value: operationResult.Status},
			{Key: "sk", Value: operationResult.SK},
			{Key: "message", Value: operationResult.Message},
			{Key: "metadata", Value: operationResult.Metadata},
			{Key: "file_data", Value: operationResult.FileData},
			{Key: "processing_date", Value: operationResult.ProcessingDate},
			{Key: "success", Value: operationResult.Success},
			{Key: "attempts", Value: operationResult.Attempts},
			{Key: "failures", Value: operationResult.Failures},
		}},
	}

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error("Error al actualizar el resultado de la operación", zap.Error(err))
		return fmt.Errorf("error al actualizar resultado de operacion: %w", err)
	}

	if result.MatchedCount == 0 {
		log.Warn("Operación no encontrada para actualizar", zap.String("id", operationID))
		return ErrOperationNotFound
	}

	log.Info("Resultado de la operación actualizado exitosamente", zap.String("id", operationID))
	return nil
}
