package dynamodb

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// OperationStore implementa la interface repository.OperationRepository maneja el almacenamiento, recuperación y eliminación de resultados de operación en DynamoDB.
type OperationStore struct {
	Client DynamoDBAPI // Cliente para interactuar con DynamoDB.
	Cfg    config.Config
}

// NewOperationStore crea una nueva instancia de OperationStore con la configuración proporcionada.
func NewOperationStore(cfgApplication config.Config) (*OperationStore, error) {
	// Carga la configuración de AWS con la región especificada.
	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(), awsCfg.WithRegion(cfgApplication.Region), awsCfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
		cfgApplication.AccessKey, cfgApplication.SecretKey, "")))
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	// Crea un nuevo cliente de DynamoDB.
	client := dynamodb.NewFromConfig(cfg)

	return &OperationStore{
		Client: client,
		Cfg:    cfgApplication,
	}, nil
}

// SaveOperationsResult guarda el resultado de una operación en DynamoDB. Genera un nuevo ID si es necesario.
func (s *OperationStore) SaveOperationsResult(ctx context.Context, result model.OperationResult) error {
	if result.ID == "" {
		result.ID = uuid.New().String() // Genera un nuevo ID si es necesario.
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(s.Cfg.OperationResultsTable),
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
		TableName: aws.String(s.Cfg.OperationResultsTable),
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
		TableName: aws.String(s.Cfg.OperationResultsTable),
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

func (s *OperationStore) UpdateOperationStatus(ctx context.Context, operationID string, songID string, status string) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(s.Cfg.OperationResultsTable),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "OPERATION#" + operationID},
			"SK": &types.AttributeValueMemberS{Value: songID},
		},
		ExpressionAttributeNames: map[string]string{
			"#status": "Status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":newStatus": &types.AttributeValueMemberS{Value: status},
		},
		UpdateExpression: aws.String("SET #status = :newStatus"),
	}

	_, err := s.Client.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("error al actualizar el estado de la operación en DynamoDB: %w", err)
	}
	return nil
}
