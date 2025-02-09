package dynamodb_test

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	dynamodb2 "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/repository/dynamodb"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	dynamodbDocker "github.com/testcontainers/testcontainers-go/modules/dynamodb"
	"testing"
	"time"
)

const (
	tableName    = "test_songs"
	partitionKey = "id"
	defaultPort  = "8000"
)

func createTestTable(ctx context.Context, client *dynamodb.Client) error {
	input := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String(partitionKey),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String(partitionKey),
				KeyType:       types.KeyTypeHash,
			},
		},
		BillingMode: types.BillingModePayPerRequest,
	}

	_, err := client.CreateTable(ctx, input)
	if err != nil {
		return err
	}

	waiter := dynamodb.NewTableExistsWaiter(client)
	return waiter.Wait(ctx, &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}, 30*time.Second)
}

func TestDynamoSongRepositoryIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Saltando prueba de integración en modo corto")
	}

	ctx := context.Background()
	logger, err := logging.NewZapLogger()
	require.NoError(t, err)

	dynamoContainer, err := dynamodbDocker.Run(ctx,
		"amazon/dynamodb-local:2.2.1",
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		if err := dynamoContainer.Terminate(ctx); err != nil {
			t.Logf("error terminando contenedor: %v", err)
		}
	})

	host, err := dynamoContainer.Host(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, host)

	mappedPort, err := dynamoContainer.MappedPort(ctx, defaultPort)
	require.NoError(t, err)

	endpoint := fmt.Sprintf("http://%s:%s", host, mappedPort.Port())

	client, err := dynamodb2.NewDynamoDBClient(ctx, "us-east-1", func(options *dynamodb.Options) {
		options.BaseEndpoint = aws.String(endpoint)
	})
	require.NoError(t, err)

	err = createTestTable(ctx, client)
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err := client.DeleteTable(ctx, &dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			t.Logf("error eliminando tabla: %v", err)
		}
	})

	repo, err := dynamodb2.NewDynamoSongRepository(dynamodb2.Options{
		TableName: tableName,
		Logger:    logger,
		Client:    client,
	})
	require.NoError(t, err)

	testSong := entity.Song{
		ID:         "song1",
		VideoID:    "video123",
		Title:      "Test Song",
		Duration:   "3:45",
		URLYoutube: "https://youtube.com/video123",
	}

	t.Run("Get Song Success", func(t *testing.T) {
		item, err := attributevalue.MarshalMap(testSong)
		require.NoError(t, err)

		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      item,
		})
		require.NoError(t, err)

		result, err := repo.GetSongByID(ctx, testSong.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, testSong.ID, result.ID)
		assert.Equal(t, testSong.Title, result.Title)
		assert.Equal(t, testSong.VideoID, result.VideoID)
		assert.Equal(t, testSong.Duration, result.Duration)
		assert.Equal(t, testSong.URLYoutube, result.URLYoutube)
	})

	t.Run("Get Non-Existent Song", func(t *testing.T) {
		result, err := repo.GetSongByID(ctx, "non-existent")
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Handle Connection Error", func(t *testing.T) {
		invalidRepo, err := dynamodb2.NewDynamoSongRepository(dynamodb2.Options{
			TableName: tableName,
			Logger:    logger,
			Client: dynamodb.NewFromConfig(aws.Config{
				Region: "invalid-region",
			}),
		})
		require.NoError(t, err)

		_, err = invalidRepo.GetSongByID(ctx, "song1")
		assert.Error(t, err)
	})
}
