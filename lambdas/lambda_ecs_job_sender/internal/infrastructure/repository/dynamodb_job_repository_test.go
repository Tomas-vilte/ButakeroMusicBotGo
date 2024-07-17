package repository

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/internal/domain/entity"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDynamoDBJobRepository_Update(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	mockLogger := new(MockLogger)
	repo := &DynamoDBJobRepository{
		client:    mockClient,
		tableName: "Jobs",
		logger:    mockLogger,
	}

	job := entity.Job{ID: "job1"}

	mockClient.On("PutItem", mock.Anything, mock.Anything).Return(&dynamodb.PutItemOutput{}, nil)
	mockLogger.On("Info", "Iniciando actualización del job", mock.Anything)
	mockLogger.On("Info", "Job actualizado exitosamente", mock.Anything)

	err := repo.Update(context.Background(), job)

	require.NoError(t, err)
	mockClient.AssertExpectations(t)

}

func TestDynamoDBJobRepository_Update_Error(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	mockLogger := new(MockLogger)
	repo := &DynamoDBJobRepository{
		client:    mockClient,
		tableName: "Jobs",
		logger:    mockLogger,
	}

	job := entity.Job{ID: "job1"}

	mockClient.On("PutItem", mock.Anything, mock.Anything).Return(&dynamodb.PutItemOutput{}, assert.AnError)
	mockLogger.On("Info", "Iniciando actualización del job", mock.Anything)
	mockLogger.On("Error", "Error al actualizar el job en DynamoDB", mock.Anything)

	err := repo.Update(context.Background(), job)

	require.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBJobRepository_GetJobByS3Key(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	mockLogger := new(MockLogger)
	repo := &DynamoDBJobRepository{
		client:    mockClient,
		tableName: "Jobs",
		logger:    mockLogger,
	}
	s3Key := "test-s3-key"
	expectedJob := entity.Job{ID: "job1"}

	mockClient.On("Query", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
		Items: []map[string]types.AttributeValue{
			{
				"ID": &types.AttributeValueMemberS{Value: expectedJob.ID},
			},
		},
	}, nil)
	mockLogger.On("Info", "Iniciando búsqueda del job por S3Key", mock.Anything)
	mockLogger.On("Info", "Job encontrado", mock.Anything)

	job, err := repo.GetJobByS3Key(context.Background(), s3Key)

	require.NoError(t, err)
	assert.Equal(t, expectedJob.ID, job.ID)
	mockClient.AssertExpectations(t)
}

func TestDynamoDBJobRepository_GetJobByS3Key_Error(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	mockLogger := new(MockLogger)
	repo := &DynamoDBJobRepository{
		client:    mockClient,
		tableName: "Jobs",
		logger:    mockLogger,
	}
	s3Key := "test-s3-key"
	mockClient.On("Query", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{}, assert.AnError)
	mockLogger.On("Info", "Iniciando búsqueda del job por S3Key", mock.Anything)
	mockLogger.On("Error", "Error al realizar la consulta a DynamoDB", mock.Anything)

	job, err := repo.GetJobByS3Key(context.Background(), s3Key)

	require.Error(t, err)
	assert.Equal(t, assert.AnError, err)
	assert.Equal(t, entity.Job{}, job)
	mockClient.AssertExpectations(t)

}
