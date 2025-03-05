package dynamodb

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// OperationStore implementa la interface repository.OperationRepository maneja el almacenamiento, recuperación y eliminación de resultados de operación en DynamoDB.
type OperationStore struct {
	Client *dynamodb.Client // Cliente para interactuar con DynamoDB.
	Cfg    *config.Config
	log    logger.Logger
}

// NewOperationStore crea una nueva instancia de OperationStore con la configuración proporcionada.
func NewOperationStore(cfgApplication *config.Config, log logger.Logger) (*OperationStore, error) {
	// Carga la configuración de AWS con la región especificada.
	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(), awsCfg.WithRegion(cfgApplication.AWS.Region))
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	// Crea un nuevo cliente de DynamoDB.
	client := dynamodb.NewFromConfig(cfg)

	return &OperationStore{
		Client: client,
		Cfg:    cfgApplication,
		log:    log,
	}, nil
}

// SaveOperationsResult guarda el resultado de una operación en DynamoDB. Genera un nuevo ID si es necesario.
func (s *OperationStore) SaveOperationsResult(ctx context.Context, result *model.OperationResult) error {
	log := s.log.With(
		zap.String("component", "OperationStore"),
		zap.String("method", "SaveOperationsResult"),
		zap.String("operation_id", result.ID),
		zap.String("song_id", result.Metadata.VideoID),
	)
	if result.ID == "" {
		result.ID = uuid.New().String()
		log.Info("Generando nuevo ID para la operación", zap.String("new_id", result.ID))
	}

	// Convierte el struct a un mapa compatible con DynamoDB
	item, err := attributevalue.MarshalMap(result)
	if err != nil {
		log.Error("Error al convertir la operación a un mapa de DynamoDB", zap.Error(err))
		return fmt.Errorf("error al convertir la operación a un mapa de DynamoDB: %w", err)
	}

	log.Info("Guardando resultado de operación en DynamoDB")
	input := &dynamodb.PutItemInput{
		TableName: aws.String(s.Cfg.Database.DynamoDB.Tables.Operations),
		Item:      item,
	}

	_, err = s.Client.PutItem(ctx, input)
	if err != nil {
		log.Error("Error al guardar resultado de operación en DynamoDB", zap.Error(err))
		return fmt.Errorf("error al guardar resultado de operación en DynamoDB: %w", err)
	}
	log.Info("Resultado de operación guardado exitosamente")
	return nil
}

// GetOperationResult recupera el resultado de una operación desde DynamoDB usando el ID y el SongID proporcionados.
func (s *OperationStore) GetOperationResult(ctx context.Context, id, songID string) (*model.OperationResult, error) {
	log := s.log.With(
		zap.String("component", "OperationStore"),
		zap.String("method", "GetOperationResult"),
		zap.String("operation_id", id),
		zap.String("song_id", songID),
	)
	input := &dynamodb.GetItemInput{
		TableName: aws.String(s.Cfg.Database.DynamoDB.Tables.Operations),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: id},
			"SK": &types.AttributeValueMemberS{Value: songID},
		},
	}

	log.Info("Obteniendo resultado de operación desde DynamoDB")
	output, err := s.Client.GetItem(ctx, input)
	if err != nil {
		log.Error("Error al recuperar resultado de operación desde DynamoDB", zap.Error(err))
		return nil, fmt.Errorf("error al recuperar resultado de operación desde DynamoDB: %w", err)
	}

	var result model.OperationResult
	if len(output.Item) == 0 {
		log.Warn("Resultado de operación no encontrado")
		return nil, fmt.Errorf("resultado de operación no encontrado")
	}

	if err := attributevalue.UnmarshalMap(output.Item, &result); err != nil {
		log.Error("Error al deserializar resultado de operación", zap.Error(err))
		return nil, fmt.Errorf("error al deserializar resultado de operación: %w", err)
	}
	log.Info("Resultado de operación obtenido exitosamente")
	return &result, nil
}

// DeleteOperationResult elimina el resultado de una operación de DynamoDB usando el ID y el SongID proporcionados.
func (s *OperationStore) DeleteOperationResult(ctx context.Context, id, songID string) error {
	log := s.log.With(
		zap.String("component", "OperationStore"),
		zap.String("method", "DeleteOperationResult"),
		zap.String("operation_id", id),
		zap.String("song_id", songID),
	)
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(s.Cfg.Database.DynamoDB.Tables.Operations),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: id},
			"SK": &types.AttributeValueMemberS{Value: songID},
		},
	}

	log.Info("Eliminando resultado de operación desde DynamoDB")
	_, err := s.Client.DeleteItem(ctx, input)
	if err != nil {
		log.Error("Error al eliminar resultado de operación desde DynamoDB", zap.Error(err))
		return fmt.Errorf("error al eliminar resultado de operación desde DynamoDB: %w", err)
	}
	log.Info("Resultado de operación eliminado exitosamente")
	return nil
}

func (s *OperationStore) UpdateOperationStatus(ctx context.Context, operationID string, songID string, status string) error {
	log := s.log.With(
		zap.String("component", "OperationStore"),
		zap.String("method", "UpdateOperationStatus"),
		zap.String("operation_id", operationID),
		zap.String("song_id", songID),
		zap.String("new_status", status),
	)
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(s.Cfg.Database.DynamoDB.Tables.Operations),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: operationID},
			"SK": &types.AttributeValueMemberS{Value: songID},
		},
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":newStatus": &types.AttributeValueMemberS{Value: status},
		},
		UpdateExpression: aws.String("SET #status = :newStatus"),
	}

	log.Info("Actualizando estado de la operación en DynamoDB")
	_, err := s.Client.UpdateItem(ctx, input)
	if err != nil {
		log.Error("Error al actualizar el estado de la operación en DynamoDB", zap.Error(err))
		return fmt.Errorf("error al actualizar el estado de la operación en DynamoDB: %w", err)
	}
	log.Info("Estado de la operación actualizado exitosamente")
	return nil
}

func (s *OperationStore) UpdateOperationResult(ctx context.Context, operationID string, operationResult *model.OperationResult) error {
	log := s.log.With(
		zap.String("component", "OperationStore"),
		zap.String("method", "UpdateOperationResult"),
		zap.String("operation_id", operationID),
		zap.String("song_id", operationResult.Metadata.VideoID),
	)
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(s.Cfg.Database.DynamoDB.Tables.Operations),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: operationID},
			"SK": &types.AttributeValueMemberS{Value: operationResult.Metadata.VideoID},
		},
		UpdateExpression: aws.String("SET #status = :status, #message = :message, #metadata = :metadata, #file_data = :file_data, #processing_date = :processing_date, #success = :success, #attempts = :attempts, #failures = :failures"),
		ExpressionAttributeNames: map[string]string{
			"#status":          "status",
			"#message":         "message",
			"#metadata":        "metadata",
			"#file_data":       "file_data",
			"#processing_date": "processing_date",
			"#success":         "success",
			"#attempts":        "attempts",
			"#failures":        "failures",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status":          &types.AttributeValueMemberS{Value: operationResult.Status},
			":message":         &types.AttributeValueMemberS{Value: operationResult.Message},
			":metadata":        &types.AttributeValueMemberM{Value: operationResult.Metadata.ToAttributeValue()},
			":file_data":       &types.AttributeValueMemberM{Value: operationResult.FileData.ToAttributeValue()},
			":processing_date": &types.AttributeValueMemberS{Value: operationResult.ProcessingDate},
			":success":         &types.AttributeValueMemberBOOL{Value: operationResult.Success},
			":attempts":        &types.AttributeValueMemberN{Value: fmt.Sprint(operationResult.Attempts)},
			":failures":        &types.AttributeValueMemberN{Value: fmt.Sprint(operationResult.Failures)},
		},
		ReturnValues: types.ReturnValueUpdatedNew,
	}

	log.Info("Actualizando resultado de operación en DynamoDB")
	_, err := s.Client.UpdateItem(ctx, input)
	if err != nil {
		log.Error("Error al actualizar resultado de operación en DynamoDB", zap.Error(err))
		return fmt.Errorf("error al actualizar resultado de operacion: %w", err)
	}

	log.Info("Resultado de operación actualizado exitosamente")
	return nil
}
