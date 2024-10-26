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

// ErrOperationNotFound es un error personalizado para cuando no se encuentra una operación
var ErrOperationNotFound = errors.New("operación no encontrada")

// OperationRepository implementa la interface port.OperationRepository para MongoDB
type (
	OperationRepository struct {
		collection *mongo.Collection
		log        logger.Logger
	}

	// OperationOptions contiene las opciones para crear un nuevo OperationRepository
	OperationOptions struct {
		Collection *mongo.Collection
		Log        logger.Logger
	}
)

// NewOperationRepository crea una nueva instancia de OperationRepository
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

// SaveOperationsResult guarda un resultado de operación en MongoDB
func (s *OperationRepository) SaveOperationsResult(ctx context.Context, result *model.OperationResult) error {
	if result == nil {
		return errors.New("result no puede ser nil")
	}

	if result.PK == "" {
		result.PK = uuid.New().String()
	}

	if _, err := s.collection.InsertOne(ctx, result); err != nil {
		s.log.Error("Error al guardar resultado de operación:", zap.Error(err))
		return fmt.Errorf("error al guardar resultado de operación: %w", err)
	}

	s.log.Info("Operacion guardada exitosamente", zap.String("id", result.PK))
	return nil
}

// GetOperationResult obtiene un resultado de operación por ID y songID
func (s *OperationRepository) GetOperationResult(ctx context.Context, id, songID string) (*model.OperationResult, error) {
	if id == "" || songID == "" {
		return nil, errors.New("id y songID son requeridos")
	}

	filter := bson.M{
		"pk": id,
		"sk": songID,
	}

	var result model.OperationResult

	if err := s.collection.FindOne(ctx, filter).Decode(&result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrOperationNotFound
		}
		s.log.Error("Error al recuperar operación:", zap.Error(err))
		return nil, fmt.Errorf("error al recuperar operación: %w", err)
	}

	return &result, nil
}

// DeleteOperationResult elimina un resultado de operación
func (s *OperationRepository) DeleteOperationResult(ctx context.Context, id, songID string) error {
	if id == "" || songID == "" {
		return errors.New("id y songID son requeridos")
	}

	filter := bson.M{
		"pk": id,
		"sk": songID,
	}

	result, err := s.collection.DeleteOne(ctx, filter)
	if err != nil {
		s.log.Error("Error al eliminar la operacion:", zap.Error(err))
		return fmt.Errorf("error al eliminar el resultado de operacion desde MongoDB: %w", err)
	}

	if result.DeletedCount == 0 {
		return ErrOperationNotFound
	}

	s.log.Info("Operacion eliminada con exito", zap.String("id", id), zap.String("songID", songID))
	return nil
}

// UpdateOperationStatus actualiza el estado de una operación
func (s *OperationRepository) UpdateOperationStatus(ctx context.Context, operationID string, songID string, status string) error {
	if operationID == "" || songID == "" || status == "" {
		return errors.New("operationID, songID y status son requeridos")
	}

	filter := bson.M{
		"pk": operationID,
		"sk": songID,
	}

	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		s.log.Error("Error al actualizar estado de operacion:", zap.Error(err))
		return fmt.Errorf("error al actualizar el estado de la operacion en MongoDB: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrOperationNotFound
	}

	s.log.Info("Estado de operacion actualizado exitosamente", zap.String("id", operationID), zap.String("status", status))
	return nil
}
