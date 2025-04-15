package dynamodb

import (
	"context"
	"errors"
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
	"strings"
)

type (
	MediaRepositoryDynamoDB struct {
		client *dynamodb.Client
		log    logger.Logger
		cfg    *config.Config
	}

	Option func(*MediaRepositoryDynamoDB)
)

func WithClient(client *dynamodb.Client) Option {
	return func(r *MediaRepositoryDynamoDB) {
		r.client = client
	}
}

func WithDefaultClient(cfgApplication *config.Config) Option {
	return func(r *MediaRepositoryDynamoDB) {
		cfg, err := awsCfg.LoadDefaultConfig(context.Background(), awsCfg.WithRegion(cfgApplication.AWS.Region))
		if err != nil {
			r.log.Error("Error cargando configuración AWS", zap.Error(err))
			return
		}
		r.client = dynamodb.NewFromConfig(cfg)
	}
}

func NewMediaRepositoryDynamoDB(cfgApplication *config.Config, log logger.Logger, opts ...Option) *MediaRepositoryDynamoDB {
	repo := &MediaRepositoryDynamoDB{
		cfg: cfgApplication,
		log: log,
	}

	for _, opt := range opts {
		opt(repo)
	}

	if repo.client == nil {
		WithDefaultClient(cfgApplication)(repo)
	}

	return repo
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

	item, err := r.toAttributeValueMap(media)
	if err != nil {
		log.Error("Error al convertir media a atributos de DynamoDB", zap.Error(err))
		return errorsApp.ErrCodeSaveMediaFailed.WithMessage(fmt.Sprintf("error al convertir media a atributos de DynamoDB: %v", err))
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(r.cfg.Database.DynamoDB.Tables.Songs),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	})
	if err != nil {
		var condCheckErr *types.ConditionalCheckFailedException
		if errors.As(err, &condCheckErr) {
			log.Warn("Registro de media ya existe en DynamoDB", zap.String("video_id", media.VideoID))
			return errorsApp.ErrDuplicateRecord.WithMessage(
				fmt.Sprintf("El video con ID '%s' ya está registrado.", media.VideoID),
				media.VideoID,
			)
		}
		log.Error("Error al guardar el registro de media en DynamoDB", zap.Error(err))
		return errorsApp.ErrCodeSaveMediaFailed.WithMessage(fmt.Sprintf("error al guardar el registro de media en DynamoDB: %v", err))
	}

	log.Info("Registro de media guardado exitosamente en DynamoDB")
	return nil
}

func (r *MediaRepositoryDynamoDB) GetMediaByID(ctx context.Context, videoID string) (*model.Media, error) {
	log := r.log.With(
		zap.String("component", "MediaRepository"),
		zap.String("method", "GetMediaByID"),
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

// GetMediaByTitle implementa la búsqueda de canciones por título en DynamoDB
func (r *MediaRepositoryDynamoDB) GetMediaByTitle(ctx context.Context, title string) ([]*model.Media, error) {
	log := r.log.With(
		zap.String("method", "SearchSongsByTitle"),
		zap.String("title", title),
	)

	normalizedTitle := strings.ToLower(title)
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.cfg.Database.DynamoDB.Tables.Songs),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :pk AND begins_with(GSI1SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "SONG"},
			":sk": &types.AttributeValueMemberS{Value: normalizedTitle},
		},
	}

	result, err := r.client.Query(ctx, input)
	if err != nil {
		log.Error("Error al buscar canciones", zap.Error(err))
		return nil, err
	}

	var songs []*model.Media
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &songs); err != nil {
		log.Error("Error de deserialización", zap.Error(err))
		return nil, err
	}

	log.Info("Búsqueda de canciones completada", zap.Int("count", len(songs)))
	return songs, nil
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
