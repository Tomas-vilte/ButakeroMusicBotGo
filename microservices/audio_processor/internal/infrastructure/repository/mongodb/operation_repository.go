// Paquete mongodb gestiona la interacción con una colección MongoDB
// para el almacenamiento y manipulación de resultados de operaciones.

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

// Errores predefinidos para condiciones específicas
var (
	ErrOperationNotFound = errors.New("operación no encontrada") // Error si la operación no se encuentra
	ErrInvalidUUID       = errors.New("UUID inválido")           // Error si un UUID no es válido
	ErrInvalidStatus     = errors.New("estado inválido")         // Error si el estado es inválido
)

// ValidStatus define los estados permitidos para una operación.
var ValidStatus = map[string]bool{
	"pending":    true,
	"complete":   true,
	"failed":     true,
	"processing": true,
}

// OperationRepository define el repositorio de operaciones con MongoDB.
type (
	OperationRepository struct {
		collection *mongo.Collection // Colección de MongoDB para almacenar operaciones
		log        logger.Logger     // Interfaz de logging
	}

	// OperationOptions agrupa las opciones necesarias para inicializar OperationRepository.
	OperationOptions struct {
		Collection *mongo.Collection // Colección de MongoDB
		Log        logger.Logger     // Logger para registrar eventos
	}
)

// isValidUUID verifica si una cadena es un UUID válido
func isValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

// createSafeFilter crea un filtro BSON para las consultas, usando IDs válidos
func createSafeFilter(pk, sk string) (bson.D, error) {
	if !isValidUUID(pk) || !isValidUUID(sk) {
		return nil, ErrInvalidUUID
	}
	// Ensure the values are properly sanitized
	sanitizedPK := bson.M{"$eq": pk}
	sanitizedSK := bson.M{"$eq": sk}
	return bson.D{
		{Key: "pk", Value: sanitizedPK},
		{Key: "sk", Value: sanitizedSK},
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
	if result == nil {
		return errors.New("result no puede ser nil")
	}
	if result.PK == "" {
		result.PK = uuid.New().String() // Genera un UUID si PK está vacío
	}

	// Validar UUIDs
	if !isValidUUID(result.PK) || !isValidUUID(result.SK) {
		return ErrInvalidUUID
	}

	// Intento de inserción en MongoDB y manejo de errores
	if _, err := s.collection.InsertOne(ctx, result); err != nil {
		s.log.Error("Error al guardar resultado de operación:", zap.Error(err))
		return fmt.Errorf("error al guardar resultado de operación: %w", err)
	}

	s.log.Info("Operacion guardada exitosamente", zap.String("id", result.PK))
	return nil
}

// GetOperationResult obtiene el resultado de una operación a partir de su ID y songID.
func (s *OperationRepository) GetOperationResult(ctx context.Context, id, songID string) (*model.OperationResult, error) {
	filter, err := createSafeFilter(id, songID)
	if err != nil {
		return nil, err
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

// DeleteOperationResult elimina el resultado de una operación específica en MongoDB.
func (s *OperationRepository) DeleteOperationResult(ctx context.Context, id, songID string) error {
	filter, err := createSafeFilter(id, songID)
	if err != nil {
		return err
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

// UpdateOperationStatus actualiza el estado de una operación, si el estado es válido.
func (s *OperationRepository) UpdateOperationStatus(ctx context.Context, operationID string, songID string, status string) error {
	if !ValidStatus[status] {
		return ErrInvalidStatus
	}

	filter, err := createSafeFilter(operationID, songID)
	if err != nil {
		return err
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "status", Value: status},
		}},
	}

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		s.log.Error("Error al actualizar estado de operacion:", zap.Error(err))
		return fmt.Errorf("error al actualizar el estado de la operacion en MongoDB: %w", err)
	}

	if result.MatchedCount == 0 {
		return ErrOperationNotFound
	}

	s.log.Info("Estado de operacion actualizado exitosamente",
		zap.String("id", operationID),
		zap.String("status", status))
	return nil
}
