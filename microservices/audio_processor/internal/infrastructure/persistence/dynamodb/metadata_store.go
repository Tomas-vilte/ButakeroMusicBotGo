package dynamodb

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

type (
	// MetadataStore Implementa la interface repository.MetadataRepository proporciona operaciones para almacenar, recuperar y eliminar metadatos en DynamoDB.
	MetadataStore struct {
		Client    DynamoDBAPI // Cliente para interactuar con DynamoDB.
		TableName string      // Nombre de la tabla en DynamoDB.
	}

	// DynamoDBAPI define los métodos necesarios para interactuar con DynamoDB.
	DynamoDBAPI interface {
		PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
		GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
		DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
		UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	}
)

// NewMetadataStore crea una nueva instancia de MetadataStore con la configuración proporcionada.
func NewMetadataStore(tableName, region string) (*MetadataStore, error) {
	// Carga la configuración de AWS con la región especificada.

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	return &MetadataStore{
		Client:    client,
		TableName: tableName,
	}, nil
}

// SaveMetadata guarda los metadatos en DynamoDB. Genera un nuevo ID si no está presente y usa la fecha actual si DownloadDate está vacío.
func (s *MetadataStore) SaveMetadata(ctx context.Context, metadata model.Metadata) error {
	if metadata.ID == "" {
		metadata.ID = uuid.New().String()
	}

	_, err := s.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.TableName),
		Item: map[string]types.AttributeValue{
			"PK":         &types.AttributeValueMemberS{Value: "METADATA#" + metadata.ID},
			"SK":         &types.AttributeValueMemberS{Value: "METADATA#" + metadata.ID},
			"ID":         &types.AttributeValueMemberS{Value: metadata.ID},
			"Title":      &types.AttributeValueMemberS{Value: metadata.Title},
			"URLYoutube": &types.AttributeValueMemberS{Value: metadata.URLYouTube},
			"Thumbnail":  &types.AttributeValueMemberS{Value: metadata.Thumbnail},
			"URLS3":      &types.AttributeValueMemberS{Value: metadata.URLS3},
			"Platform":   &types.AttributeValueMemberS{Value: metadata.Platform},
			"Duration":   &types.AttributeValueMemberS{Value: metadata.Duration},
		},
	})
	if err != nil {
		return fmt.Errorf("error al guardar resultado de operación en DynamoDB: %w", err)
	}
	return nil
}

// GetMetadata recupera los metadatos de DynamoDB usando el ID proporcionado.
func (s *MetadataStore) GetMetadata(ctx context.Context, id string) (*model.Metadata, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(s.TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "METADATA#" + id},
			"SK": &types.AttributeValueMemberS{Value: "METADATA#" + id},
		},
	}

	output, err := s.Client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error al recuperar metadatos desde DynamoDB: %w", err)
	}
	var metadata model.Metadata
	if len(output.Item) == 0 {
		return nil, fmt.Errorf("metadatos no encontrados")
	}

	if err := attributevalue.UnmarshalMap(output.Item, &metadata); err != nil {
		return nil, fmt.Errorf("error al deserializar metadatos: %w", err)
	}
	return &metadata, nil
}

// DeleteMetadata elimina los metadatos de DynamoDB usando el ID proporcionado.
func (s *MetadataStore) DeleteMetadata(ctx context.Context, id string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(s.TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "METADATA#" + id},
			"SK": &types.AttributeValueMemberS{Value: "METADATA#" + id},
		},
	}
	_, err := s.Client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("error al eliminar metadatos desde DynamoDB: %w", err)
	}
	return nil
}