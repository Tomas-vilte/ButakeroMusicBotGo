package dynamodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.uber.org/zap"
	"strings"
)

var (
	ErrInvalidConfig = errors.New("configuracion invalida")
)

type (
	// Options contiene todas las opciones de configuración del repositorio
	Options struct {
		TableName string
		Logger    logging.Logger
		Client    *dynamodb.Client
	}

	// DynamoSongRepository implementa la interfaz ports.SongRepository para DynamoDB
	DynamoSongRepository struct {
		opts Options
	}
)

// validate realiza validaciones básicas de la configuración
func (o *Options) validate() error {
	if o.Client == nil && o.TableName == "" {
		return fmt.Errorf("%w client y tableName requeridos", ErrInvalidConfig)
	}

	if o.Logger == nil {
		return fmt.Errorf("%w logger es requerido", ErrInvalidConfig)
	}
	return nil
}

// NewDynamoSongRepository crea un nuevo repositorio con opciones configuradas
func NewDynamoSongRepository(opts Options) (*DynamoSongRepository, error) {
	if err := opts.validate(); err != nil {
		return nil, fmt.Errorf("error validando opciones: %w", err)
	}

	return &DynamoSongRepository{
		opts: opts,
	}, nil
}

// GetSongByVideoID implementa la obtención de canciones desde DynamoDB
func (r *DynamoSongRepository) GetSongByVideoID(ctx context.Context, videoID string) (*entity.Song, error) {
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.opts.TableName),
		IndexName:              aws.String("VideoIDIndex"),
		KeyConditionExpression: aws.String("video_id = :videoID"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":videoID": &types.AttributeValueMemberS{Value: videoID},
		},
	}

	result, err := r.opts.Client.Query(ctx, input)
	if err != nil {
		r.opts.Logger.Error("Error al obtener cancion",
			zap.String("id", videoID),
			zap.Error(err),
		)
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, nil
	}

	var song entity.Song
	if err := attributevalue.UnmarshalMap(result.Items[0], &song); err != nil {
		r.opts.Logger.Error("Error de deserialzacion",
			zap.Error(err))
		return nil, fmt.Errorf("error deserializando canción: %w", err)
	}
	return &song, nil
}

// SearchSongsByTitle implementa la búsqueda de canciones por título en DynamoDB
func (r *DynamoSongRepository) SearchSongsByTitle(ctx context.Context, title string) ([]*entity.Song, error) {
	normalizedTitle := strings.ToLower(title)
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.opts.TableName),
		IndexName:              aws.String("GSI2-title-index"),
		KeyConditionExpression: aws.String("GSI2_PK = :pk AND begins_with(title, :title)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":    &types.AttributeValueMemberS{Value: "SEARCH#TITLE"},
			":title": &types.AttributeValueMemberS{Value: normalizedTitle},
		},
	}

	result, err := r.opts.Client.Query(ctx, input)
	if err != nil {
		r.opts.Logger.Error("Error al buscar canciones",
			zap.String("title", title),
			zap.Error(err),
		)
		return nil, err
	}

	var songs []*entity.Song
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &songs); err != nil {
		r.opts.Logger.Error("Error de deserialzacion",
			zap.Error(err))
		return nil, err
	}

	return songs, nil
}
