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
func (r *DynamoSongRepository) GetSongByVideoID(ctx context.Context, videoID string) (*entity.SongEntity, error) {
	logger := r.opts.Logger.With(
		zap.String("method", "GetSongByVideoID"),
		zap.String("videoID", videoID),
	)

	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.opts.TableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "VIDEO#" + videoID},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		},
	}

	result, err := r.opts.Client.GetItem(ctx, input)
	if err != nil {
		logger.Error("Error al obtener canción", zap.Error(err))
		return nil, err
	}

	if len(result.Item) == 0 {
		logger.Info("Canción no encontrada")
		return nil, nil
	}

	var song entity.SongEntity
	if err := attributevalue.UnmarshalMap(result.Item, &song); err != nil {
		logger.Error("Error de deserialización", zap.Error(err))
		return nil, fmt.Errorf("error deserializando canción: %w", err)
	}

	logger.Info("Canción obtenida exitosamente")
	return &song, nil
}

// SearchSongsByTitle implementa la búsqueda de canciones por título en DynamoDB
func (r *DynamoSongRepository) SearchSongsByTitle(ctx context.Context, title string) ([]*entity.SongEntity, error) {
	logger := r.opts.Logger.With(
		zap.String("method", "SearchSongsByTitle"),
		zap.String("title", title),
	)

	normalizedTitle := strings.ToLower(title)
	input := &dynamodb.QueryInput{
		TableName:              aws.String(r.opts.TableName),
		IndexName:              aws.String("GSI1"),
		KeyConditionExpression: aws.String("GSI1PK = :pk AND begins_with(GSI1SK, :sk)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: "SONG"},
			":sk": &types.AttributeValueMemberS{Value: normalizedTitle},
		},
	}

	result, err := r.opts.Client.Query(ctx, input)
	if err != nil {
		logger.Error("Error al buscar canciones", zap.Error(err))
		return nil, err
	}

	var songs []*entity.SongEntity
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &songs); err != nil {
		logger.Error("Error de deserialización", zap.Error(err))
		return nil, err
	}

	logger.Info("Búsqueda de canciones completada", zap.Int("count", len(songs)))
	return songs, nil
}
