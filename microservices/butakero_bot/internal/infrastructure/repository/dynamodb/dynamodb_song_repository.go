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

// GetSongByID implementa la obtención de canciones desde DynamoDB
func (r *DynamoSongRepository) GetSongByID(ctx context.Context, id string) (*entity.Song, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.opts.TableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	}

	result, err := r.opts.Client.GetItem(ctx, input)
	if err != nil {
		r.opts.Logger.Error("Error al obtener cancion",
			zap.String("id", id),
			zap.Error(err),
		)
		return nil, fmt.Errorf("error al obtener cancion: %w", err)
	}

	if result.Item == nil {
		r.opts.Logger.Info("Cancion no encontrada", zap.String("id", id))
		return nil, nil
	}

	var song entity.Song
	if err := attributevalue.UnmarshalMap(result.Item, &song); err != nil {
		r.opts.Logger.Error("Error de deserialzacion",
			zap.String("id", id),
			zap.Error(err))
		return nil, fmt.Errorf("error deserializando canción: %w", err)
	}

	return &song, nil
}
