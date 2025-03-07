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
	"strings"
)

type (
	// DynamoMetadataRepository Implementa la interface repository.MetadataRepository proporciona operaciones para almacenar, recuperar y eliminar metadatos en DynamoDB.
	DynamoMetadataRepository struct {
		Client *dynamodb.Client // Cliente para interactuar con DynamoDB.
		Config *config.Config
		log    logger.Logger
	}
)

// NewMetadataStore crea una nueva instancia de MetadataStore con la configuración proporcionada.
func NewMetadataStore(cfgApplication *config.Config, log logger.Logger) (*DynamoMetadataRepository, error) {
	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(), awsCfg.WithRegion(cfgApplication.AWS.Region))
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	return &DynamoMetadataRepository{
		Client: client,
		Config: cfgApplication,
		log:    log,
	}, nil
}

// SaveMetadata guarda los metadatos en DynamoDB. Genera un nuevo ID si no está presente y usa la fecha actual si DownloadDate está vacío.
func (s *DynamoMetadataRepository) SaveMetadata(ctx context.Context, metadata *model.Metadata) error {
	log := s.log.With(
		zap.String("component", "DynamoMetadataRepository"),
		zap.String("method", "SaveMetadata"),
		zap.String("metadata_id", metadata.ID),
	)
	if metadata.ID == "" {
		metadata.ID = uuid.New().String()
		log.Info("Generando nuevo ID para los metadatos", zap.String("new_id", metadata.ID))
	}

	metadata.Title = strings.ToLower(metadata.Title)

	_, err := s.Client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(s.Config.Database.DynamoDB.Tables.Songs),
		Item: map[string]types.AttributeValue{
			"PK":            &types.AttributeValueMemberS{Value: "METADATA#" + metadata.ID},
			"SK":            &types.AttributeValueMemberS{Value: "METADATA#" + metadata.ID},
			"GSI2_PK":       &types.AttributeValueMemberS{Value: "SEARCH#TITLE"},
			"ID":            &types.AttributeValueMemberS{Value: metadata.ID},
			"title":         &types.AttributeValueMemberS{Value: metadata.Title},
			"url":           &types.AttributeValueMemberS{Value: metadata.URL},
			"thumbnail_url": &types.AttributeValueMemberS{Value: metadata.ThumbnailURL},
			"platform":      &types.AttributeValueMemberS{Value: metadata.Platform},
			"duration_ms":   &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", metadata.DurationMs)},
		},
	})
	if err != nil {
		log.Error("Error al guardar metadatos en DynamoDB", zap.Error(err))
		return fmt.Errorf("error al guardar resultado de operación en DynamoDB: %w", err)
	}
	return nil
}

// GetMetadata recupera los metadatos de DynamoDB usando el ID proporcionado.
func (s *DynamoMetadataRepository) GetMetadata(ctx context.Context, id string) (*model.Metadata, error) {
	log := s.log.With(
		zap.String("component", "DynamoMetadataRepository"),
		zap.String("method", "GetMetadata"),
		zap.String("metadata_id", id),
	)

	input := &dynamodb.GetItemInput{
		TableName: aws.String(s.Config.Database.DynamoDB.Tables.Songs),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "METADATA#" + id},
			"SK": &types.AttributeValueMemberS{Value: "METADATA#" + id},
		},
	}

	log.Info("Obteniendo metadatos desde DynamoDB")
	output, err := s.Client.GetItem(ctx, input)
	if err != nil {
		log.Error("Error al recuperar metadatos desde DynamoDB", zap.Error(err))
		return nil, fmt.Errorf("error al recuperar metadatos desde DynamoDB: %w", err)
	}
	var metadata model.Metadata
	if len(output.Item) == 0 {
		log.Warn("Metadatos no encontrados")
		return nil, fmt.Errorf("metadatos no encontrados")
	}

	if err := attributevalue.UnmarshalMap(output.Item, &metadata); err != nil {
		log.Error("Error al deserializar metadatos", zap.Error(err))
		return nil, fmt.Errorf("error al deserializar metadatos: %w", err)
	}
	log.Info("Metadatos obtenidos exitosamente")
	return &metadata, nil
}

// DeleteMetadata elimina los metadatos de DynamoDB usando el ID proporcionado.
func (s *DynamoMetadataRepository) DeleteMetadata(ctx context.Context, id string) error {
	log := s.log.With(
		zap.String("component", "DynamoMetadataRepository"),
		zap.String("method", "DeleteMetadata"),
		zap.String("metadata_id", id),
	)
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(s.Config.Database.DynamoDB.Tables.Songs),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "METADATA#" + id},
			"SK": &types.AttributeValueMemberS{Value: "METADATA#" + id},
		},
	}

	log.Info("Eliminando metadatos desde DynamoDB")
	_, err := s.Client.DeleteItem(ctx, input)
	if err != nil {
		log.Error("Error al eliminar metadatos desde DynamoDB", zap.Error(err))
		return fmt.Errorf("error al eliminar metadatos desde DynamoDB: %w", err)
	}
	log.Info("Metadatos eliminados exitosamente")
	return nil
}
