package repository

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/internal/domain/entity"
	logging "github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/pkg/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.uber.org/zap"
)

type (
	// DynamoDBJobRepository es una implementación del JobRepository que usa DynamoDB.
	DynamoDBJobRepository struct {
		client    DynamoDBClient
		tableName string
		logger    logging.Logger
	}

	// DynamoDBClient es una interfaz que define los métodos del cliente de DynamoDB que se van a usar.
	DynamoDBClient interface {
		PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
		Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	}
)

// NewDynamoDBJobRepository crea una nueva instancia de DynamoDBJobRepository.
func NewDynamoDBJobRepository(logger logging.Logger) (*DynamoDBJobRepository, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := dynamodb.NewFromConfig(cfg)

	return &DynamoDBJobRepository{
		client:    client,
		tableName: "Jobs",
		logger:    logger,
	}, nil
}

// Update actualiza un job existente en la tabla DynamoDB.
func (r *DynamoDBJobRepository) Update(ctx context.Context, job entity.Job) error {
	r.logger.Info("Iniciando actualización del job", zap.String("jobID", job.ID))

	item, err := attributevalue.MarshalMap(job)
	if err != nil {
		r.logger.Error("Error al convertire el job a un mapa", zap.Error(err))
		return err
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	})
	if err != nil {
		r.logger.Error("Error al actualizar el job en DynamoDB", zap.Error(err))
		return err
	}

	r.logger.Info("Job actualizado exitosamente", zap.String("jobID", job.ID))
	return err
}

func (r *DynamoDBJobRepository) GetJobByS3Key(ctx context.Context, s3Key string) (entity.Job, error) {
	r.logger.Info("Iniciando búsqueda del job por S3Key", zap.String("S3Key", s3Key))

	result, err := r.client.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(r.tableName),
		IndexName: aws.String("S3Key-index"),
		KeyConditions: map[string]types.Condition{
			"KEY": {
				ComparisonOperator: types.ComparisonOperatorEq,
				AttributeValueList: []types.AttributeValue{
					&types.AttributeValueMemberS{Value: s3Key},
				},
			},
		},
	})
	if err != nil {
		r.logger.Error("Error al realizar la consulta a DynamoDB", zap.Error(err))
		return entity.Job{}, err
	}

	if len(result.Items) == 0 {
		r.logger.Info("No se encontro ningun job con el S3Key", zap.String("S3Key", s3Key))
		return entity.Job{}, nil
	}
	var job entity.Job
	err = attributevalue.UnmarshalMap(result.Items[0], &job)
	if err != nil {
		r.logger.Error("Error al deserializar el job desde DynamoDB", zap.Error(err))
		return entity.Job{}, err
	}
	r.logger.Info("Job encontrado", zap.String("jobID", job.ID))
	return job, nil
}
