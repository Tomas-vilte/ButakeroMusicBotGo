package dynamodbservice

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// OperationStore maneja el almacenamiento, recuperación y eliminación de resultados de operación en DynamoDB.
type OperationStore struct {
	Client    DynamoDBAPI // Cliente para interactuar con DynamoDB.
	TableName string      // Nombre de la tabla en DynamoDB.
}

// NewOperationStore crea una nueva instancia de OperationStore con la configuración proporcionada.
func NewOperationStore(tableName string, region string) (*OperationStore, error) {
	// Carga la configuración de AWS con la región especificada.
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	// Crea un nuevo cliente de DynamoDB.
	client := dynamodb.NewFromConfig(cfg)

	return &OperationStore{
		Client:    client,
		TableName: tableName,
	}, nil
}

// SaveOperationResult guarda el resultado de una operación en DynamoDB. Genera un nuevo ID si es necesario.
func (s *OperationStore) SaveOperationResult(ctx context.Context, result model.OperationResult) error {
	if result.ID == "" {
		result.ID = uuid.New().String() // Genera un nuevo ID si es necesario.
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(s.TableName),
		Item: map[string]types.AttributeValue{
			"PK":             &types.AttributeValueMemberS{Value: "OPERATION#" + result.ID},
			"SK":             &types.AttributeValueMemberS{Value: result.SongID},
			"ID":             &types.AttributeValueMemberS{Value: result.ID},
			"SongID":         &types.AttributeValueMemberS{Value: result.SongID},
			"Status":         &types.AttributeValueMemberS{Value: result.Status},
			"Message":        &types.AttributeValueMemberS{Value: result.Message},
			"Data":           &types.AttributeValueMemberS{Value: result.Data},
			"ProcessingDate": &types.AttributeValueMemberS{Value: result.ProcessingDate},
			"Success":        &types.AttributeValueMemberBOOL{Value: result.Success},
			"Attempts":       &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", result.Attempts)},
			"Failures":       &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", result.Failures)},
		},
	}

	_, err := s.Client.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("error al guardar resultado de operación en DynamoDB: %w", err)
	}
	return nil
}

// GetOperationResult recupera el resultado de una operación desde DynamoDB usando el ID y el SongID proporcionados.
func (s *OperationStore) GetOperationResult(ctx context.Context, id, songID string) (*model.OperationResult, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(s.TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "OPERATION#" + id},
			"SK": &types.AttributeValueMemberS{Value: songID},
		},
	}

	output, err := s.Client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error al recuperar resultado de operación desde DynamoDB: %w", err)
	}

	var result model.OperationResult
	if len(output.Item) == 0 {
		return nil, fmt.Errorf("resultado de operación no encontrado")
	}

	if err := attributevalue.UnmarshalMap(output.Item, &result); err != nil {
		return nil, fmt.Errorf("error al deserializar resultado de operación: %w", err)
	}
	return &result, nil
}

// DeleteOperationResult elimina el resultado de una operación de DynamoDB usando el ID y el SongID proporcionados.
func (s *OperationStore) DeleteOperationResult(ctx context.Context, id, songID string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(s.TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "OPERATION#" + id},
			"SK": &types.AttributeValueMemberS{Value: songID},
		},
	}
	_, err := s.Client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("error al eliminar resultado de operación desde DynamoDB: %w", err)
	}
	return nil
}
