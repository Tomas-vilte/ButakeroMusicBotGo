package dynamodb

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	errorsApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.uber.org/zap"
	"time"
)

type (
	MediaRepositoryDynamoDB struct {
		client *dynamodb.Client
		log    logger.Logger
		cfg    *config.Config
	}
)

func NewMediaRepositoryDynamoDB(cfgApplication *config.Config, log logger.Logger) (*MediaRepositoryDynamoDB, error) {
	cfg, err := awsCfg.LoadDefaultConfig(context.Background(), awsCfg.WithRegion(cfgApplication.AWS.Region))
	if err != nil {
		return nil, errorsApp.ErrCodeDBConnectionFailed.WithMessage(fmt.Sprintf("error cargando configuración AWS: %v", err))
	}

	client := dynamodb.NewFromConfig(cfg)

	return &MediaRepositoryDynamoDB{
		cfg:    cfgApplication,
		client: client,
		log:    log,
	}, nil
}

func (r *MediaRepositoryDynamoDB) SaveMedia(ctx context.Context, media *model.Media) error {
	log := r.log.With(
		zap.String("component", "MediaRepository"),
		zap.String("method", "SaveMedia"),
	)

	media.PK = fmt.Sprintf("VIDEO#%s", media.VideoID)
	media.SK = "METADATA"
	media.GSI1PK = "SONG"
	media.GSI1SK = media.TitleLower

	now := time.Now()
	media.CreatedAt = now
	media.UpdatedAt = now

	item, err := r.toAttributeValueMap(media)
	if err != nil {
		log.Error("Error al convertir media a atributos de DynamoDB", zap.Error(err))
		return errorsApp.ErrCodeSaveMediaFailed.WithMessage(fmt.Sprintf("error al convertir media a atributos de DynamoDB: %v", err))
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.cfg.Database.DynamoDB.Tables.Songs),
		Item:      item,
	})
	if err != nil {
		log.Error("Error al guardar el registro de media en DynamoDB", zap.Error(err))
		return errorsApp.ErrCodeSaveMediaFailed.WithMessage(fmt.Sprintf("error al guardar el registro de media en DynamoDB: %v", err))
	}

	log.Info("Registro de media guardado exitosamente en DynamoDB")
	return nil
}

func (r *MediaRepositoryDynamoDB) GetMedia(ctx context.Context, videoID string) (*model.Media, error) {
	log := r.log.With(
		zap.String("component", "MediaRepository"),
		zap.String("method", "GetMedia"),
		zap.String("video_id", videoID),
	)

	result, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.cfg.Database.DynamoDB.Tables.Songs),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("VIDEO#%s", videoID)},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	})
	if err != nil {
		log.Error("Error al obtener el registro de media de DynamoDB", zap.Error(err))
		return nil, errorsApp.ErrGetMediaDetailsFailed.WithMessage(fmt.Sprintf("error al obtener el registro de media de DynamoDB: %v", err))
	}

	if result.Item == nil {
		log.Warn("Registro de media no encontrado en DynamoDB")
		return nil, errorsApp.ErrCodeMediaNotFound
	}

	media, err := r.fromAttributeValueMap(result.Item)
	if err != nil {
		log.Error("Error al convertir atributos de DynamoDB a media", zap.Error(err))
		return nil, errorsApp.ErrGetMediaDetailsFailed.WithMessage(fmt.Sprintf("error al convertir atributos de DynamoDB a media: %v", err))
	}

	media.VideoID = videoID

	log.Info("Registro de media recuperado exitosamente de DynamoDB")
	return media, nil
}

func (r *MediaRepositoryDynamoDB) DeleteMedia(ctx context.Context, videoID string) error {
	log := r.log.With(
		zap.String("component", "MediaRepository"),
		zap.String("method", "DeleteMedia"),
		zap.String("video_id", videoID),
	)

	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.cfg.Database.DynamoDB.Tables.Songs),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: fmt.Sprintf("VIDEO#%s", videoID)},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	})
	if err != nil {
		log.Error("Error al eliminar el registro de media de DynamoDB", zap.Error(err))
		return errorsApp.ErrCodeDeleteMediaFailed.WithMessage(fmt.Sprintf("error al eliminar el registro de media de DynamoDB: %v", err))
	}

	log.Info("Registro de media eliminado exitosamente de DynamoDB")
	return nil
}

func (r *MediaRepositoryDynamoDB) UpdateMedia(ctx context.Context, videoID string, media *model.Media) error {
	log := r.log.With(
		zap.String("component", "MediaRepository"),
		zap.String("method", "UpdateMedia"),
		zap.String("video_id", videoID),
	)
	media.PK = fmt.Sprintf("VIDEO#%s", media.VideoID)
	media.SK = "METADATA"
	media.GSI1PK = "SONG"
	media.GSI1SK = media.TitleLower

	item, err := r.toAttributeValueMap(media)
	if err != nil {
		log.Error("Error al convertir media a atributos de DynamoDB", zap.Error(err))
		return errorsApp.ErrUpdateMediaFailed.WithMessage(fmt.Sprintf("error al convertir media a atributos de DynamoDB: %v", err))
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.cfg.Database.DynamoDB.Tables.Songs),
		Item:      item,
	})
	if err != nil {
		log.Error("Error al actualizar el registro de media en DynamoDB", zap.Error(err))
		return errorsApp.ErrUpdateMediaFailed.WithMessage(fmt.Sprintf("error al actualizar el registro de media en DynamoDB: %v", err))
	}

	log.Info("Registro de media actualizado exitosamente en DynamoDB")
	return nil
}

func (r *MediaRepositoryDynamoDB) toAttributeValueMap(media *model.Media) (map[string]types.AttributeValue, error) {
	item, err := attributevalue.MarshalMap(media)
	if err != nil {
		return nil, fmt.Errorf("error al convertir media a atributos de DynamoDB: %w", err)
	}
	return item, nil
}

func (r *MediaRepositoryDynamoDB) fromAttributeValueMap(item map[string]types.AttributeValue) (*model.Media, error) {
	var media model.Media
	err := attributevalue.UnmarshalMap(item, &media)
	if err != nil {
		return nil, fmt.Errorf("error al convertir atributos de DynamoDB a media: %w", err)
	}
	return &media, nil
}
