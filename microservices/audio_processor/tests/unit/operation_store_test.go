package unit

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	dynamodb2 "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/persistence/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestSaveOperationResult(t *testing.T) {
	t.Run("Successful save", func(t *testing.T) {
		mockClient := new(MockDynamoDBAPI)
		store := dynamodb2.OperationStore{
			Client: mockClient,
			Cfg: config.Config{
				OperationResultsTable: "TestOperationStore",
			},
		}

		result := model.OperationResult{
			SK: "test-song-id",
		}

		mockClient.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput"), mock.Anything).
			Return(&dynamodb.PutItemOutput{}, nil)

		err := store.SaveOperationsResult(context.Background(), result)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("DynamoDB error", func(t *testing.T) {
		mockClient := new(MockDynamoDBAPI)
		store := dynamodb2.OperationStore{
			Client: mockClient,
			Cfg: config.Config{
				OperationResultsTable: "TestOperationStore",
			},
		}

		result := model.OperationResult{
			SK: "test-song-id",
		}

		mockClient.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput"), mock.Anything).
			Return((*dynamodb.PutItemOutput)(nil), errors.New("dynamoDB error"))

		err := store.SaveOperationsResult(context.Background(), result)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error al guardar resultado de operación en DynamoDB")
		mockClient.AssertExpectations(t)
	})
}

func TestGetOperationResult(t *testing.T) {
	t.Run("Successful retrieval", func(t *testing.T) {
		mockClient := new(MockDynamoDBAPI)
		store := dynamodb2.OperationStore{
			Client: mockClient,
			Cfg: config.Config{
				OperationResultsTable: "TestOperationStore",
			},
		}

		expectedResult := &model.OperationResult{
			SK: "test-song-id",
		}

		mockClient.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).
			Return(&dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"SK": &types.AttributeValueMemberS{Value: "test-song-id"},
				},
			}, nil)

		result, err := store.GetOperationResult(context.Background(), "test-operation-id", "test-song-id-123")

		assert.NoError(t, err)
		assert.Equal(t, expectedResult, result)
		mockClient.AssertExpectations(t)

	})

	t.Run("Item not found", func(t *testing.T) {
		mockClient := new(MockDynamoDBAPI)
		store := dynamodb2.OperationStore{
			Client: mockClient,
			Cfg: config.Config{
				OperationResultsTable: "TestOperationStore",
			},
		}

		mockClient.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).
			Return(&dynamodb.GetItemOutput{}, nil)

		result, err := store.GetOperationResult(context.Background(), "non-existent-operation-id", "non-existent-song-id-123")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "resultado de operación no encontrado")
		mockClient.AssertExpectations(t)
	})

	t.Run("DynamoDB error", func(t *testing.T) {
		mockClient := new(MockDynamoDBAPI)
		store := dynamodb2.OperationStore{
			Client: mockClient,
			Cfg: config.Config{
				OperationResultsTable: "TestOperationStore",
			},
		}

		mockClient.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).
			Return((*dynamodb.GetItemOutput)(nil), errors.New("dynamoDB error"))

		result, err := store.GetOperationResult(context.Background(), "test-operation-id", "test-song-id-123")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "error al recuperar resultado de operación desde DynamoDB")
		mockClient.AssertExpectations(t)
	})
}

func TestDeleteOperationResult(t *testing.T) {
	t.Run("Successful deletion", func(t *testing.T) {
		mockClient := new(MockDynamoDBAPI)
		store := dynamodb2.OperationStore{
			Client: mockClient,
			Cfg: config.Config{
				OperationResultsTable: "TestOperationStore",
			},
		}

		mockClient.On("DeleteItem", mock.Anything, mock.AnythingOfType("*dynamodb.DeleteItemInput"), mock.Anything).
			Return(&dynamodb.DeleteItemOutput{}, nil)

		err := store.DeleteOperationResult(context.Background(), "test-operation-id", "test-song-id-123")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("DynamoDB error", func(t *testing.T) {
		mockClient := new(MockDynamoDBAPI)
		store := dynamodb2.OperationStore{
			Client: mockClient,
			Cfg: config.Config{
				OperationResultsTable: "TestOperationStore",
			},
		}

		mockClient.On("DeleteItem", mock.Anything, mock.AnythingOfType("*dynamodb.DeleteItemInput"), mock.Anything).
			Return((*dynamodb.DeleteItemOutput)(nil), errors.New("dynamoDB error"))

		err := store.DeleteOperationResult(context.Background(), "test-operation-id", "test-song-id-123")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error al eliminar resultado de operación desde DynamoDB")
		mockClient.AssertExpectations(t)
	})
}

func TestUpdateOperationStatus(t *testing.T) {
	t.Run("Successful update", func(t *testing.T) {
		mockClient := new(MockDynamoDBAPI)
		store := dynamodb2.OperationStore{
			Client: mockClient,
			Cfg: config.Config{
				OperationResultsTable: "TestOperationStore",
			},
		}

		mockClient.On("UpdateItem", mock.Anything, mock.AnythingOfType("*dynamodb.UpdateItemInput"), mock.Anything).
			Return(&dynamodb.UpdateItemOutput{}, nil)

		err := store.UpdateOperationStatus(context.Background(), "test-operation-id", "test-song-id", "completed")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("DynamoDB error", func(t *testing.T) {
		mockClient := new(MockDynamoDBAPI)
		store := dynamodb2.OperationStore{
			Client: mockClient,
			Cfg: config.Config{
				OperationResultsTable: "TestOperationStore",
			},
		}

		mockClient.On("UpdateItem", mock.Anything, mock.AnythingOfType("*dynamodb.UpdateItemInput"), mock.Anything).
			Return((*dynamodb.UpdateItemOutput)(nil), errors.New("dynamoDB error"))

		err := store.UpdateOperationStatus(context.Background(), "test-operation-id", "test-song-id", "failed")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error al actualizar el estado de la operación en DynamoDB")
		mockClient.AssertExpectations(t)
	})
}
