package unit

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/dynamodbservice"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockDynamoDBAPI struct {
	mock.Mock
}

func (m *mockDynamoDBAPI) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *mockDynamoDBAPI) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *mockDynamoDBAPI) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.DeleteItemOutput), args.Error(1)
}

func TestMetadataStore(t *testing.T) {
	t.Run("SaveMetadata", func(t *testing.T) {
		t.Run("Successful save", func(t *testing.T) {
			// arrange
			mockClient := new(mockDynamoDBAPI)
			store := dynamodbservice.MetadataStore{
				Client:    mockClient,
				TableName: "test-table",
			}
			metadata := model.Metadata{
				ID:       "test-id",
				Title:    "Test Song",
				Artist:   "Test Artist",
				Duration: 180,
			}
			mockClient.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput"), mock.Anything).Return(&dynamodb.PutItemOutput{}, nil)

			// act
			err := store.SaveMetadata(context.Background(), metadata)

			// assert
			assert.NoError(t, err)
			mockClient.AssertExpectations(t)
		})
	})

	t.Run("DynamoDB error", func(t *testing.T) {
		// arrange
		mockClient := new(mockDynamoDBAPI)
		store := &dynamodbservice.MetadataStore{
			Client:    mockClient,
			TableName: "test-table",
		}
		metadata := model.Metadata{
			ID:       "test-id",
			Title:    "Test Song",
			Artist:   "Test Artist",
			Duration: 180,
		}
		mockClient.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput"), mock.Anything).Return(&dynamodb.PutItemOutput{}, errors.New("DynamoDB error"))

		// act
		err := store.SaveMetadata(context.Background(), metadata)

		// assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error al guardar resultado de operación en DynamoDB")
		mockClient.AssertExpectations(t)
	})

	t.Run("NewMetadataStore", func(t *testing.T) {
		t.Run("Successful creation", func(t *testing.T) {
			// arrange
			tableName := "test-table"
			region := "us-east-1"

			// act
			store, err := dynamodbservice.NewMetadataStore(tableName, region)

			// assert
			assert.NoError(t, err)
			assert.NotNil(t, store)
			assert.Equal(t, tableName, store.TableName)
			assert.NotNil(t, store.Client)
		})
	})

	t.Run("Successful retrieval", func(t *testing.T) {
		mockClient := new(mockDynamoDBAPI)
		store := &dynamodbservice.MetadataStore{
			Client:    mockClient,
			TableName: "TestTable",
		}
		expectedMetadata := model.Metadata{
			ID:    "test-id",
			Title: "Test Song",
		}
		marshalledItem, _ := attributevalue.MarshalMap(expectedMetadata)
		mockClient.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(&dynamodb.GetItemOutput{
			Item: marshalledItem,
		}, nil)

		result, err := store.GetMetadata(context.Background(), "test-id")
		assert.NoError(t, err)
		assert.Equal(t, &expectedMetadata, result)
		mockClient.AssertExpectations(t)
	})

	t.Run("Item not found", func(t *testing.T) {
		mockClient := new(mockDynamoDBAPI)
		store := &dynamodbservice.MetadataStore{
			Client:    mockClient,
			TableName: "TestTable",
		}
		mockClient.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)

		result, err := store.GetMetadata(context.Background(), "non-existent-id")
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "metadatos no encontrados")
		mockClient.AssertExpectations(t)
	})

	t.Run("DynamoDB error", func(t *testing.T) {
		mockClient := new(mockDynamoDBAPI)
		store := &dynamodbservice.MetadataStore{
			Client:    mockClient,
			TableName: "TestTable",
		}

		mockClient.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(
			(*dynamodb.GetItemOutput)(nil), errors.New("DynamoDB error"))

		result, err := store.GetMetadata(context.Background(), "test-id")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "error al recuperar metadatos desde DynamoDB")
		mockClient.AssertExpectations(t)
	})
}

func TestGetMetadata(t *testing.T) {
	t.Run("Successful deletion", func(t *testing.T) {
		mockClient := new(mockDynamoDBAPI)
		store := &dynamodbservice.MetadataStore{
			Client:    mockClient,
			TableName: "TestTable",
		}
		mockClient.On("DeleteItem", mock.Anything, mock.AnythingOfType("*dynamodb.DeleteItemInput"), mock.Anything).Return(&dynamodb.DeleteItemOutput{}, nil)

		err := store.DeleteMetadata(context.Background(), "test-id")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("DynamoDB error", func(t *testing.T) {
		mockClient := new(mockDynamoDBAPI)
		store := &dynamodbservice.MetadataStore{
			Client:    mockClient,
			TableName: "TestTable",
		}
		mockClient.On("DeleteItem", mock.Anything, mock.AnythingOfType("*dynamodb.DeleteItemInput"), mock.Anything).
			Return((*dynamodb.DeleteItemOutput)(nil), errors.New("DynamoDB error"))

		err := store.DeleteMetadata(context.Background(), "test-id")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error al eliminar metadatos desde DynamoDB")
		mockClient.AssertExpectations(t)
	})
}
