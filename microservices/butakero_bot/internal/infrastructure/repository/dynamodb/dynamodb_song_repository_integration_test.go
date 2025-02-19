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

func createTestTableWithIndexes(ctx context.Context, client *dynamodb.Client) error {
	input := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String(partitionKey),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("video_id"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("title"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI2_PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String(partitionKey),
				KeyType:       types.KeyTypeHash,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("VideoIDIndex"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("video_id"),
						KeyType:       types.KeyTypeHash,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
			},
			{
				IndexName: aws.String("GSI2-title-index"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("GSI2_PK"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("title"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
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
		"amazon/dynamodb-local:2.5.4",
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

	err = createTestTableWithIndexes(ctx, client)
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

	testSongs := []entity.Song{
		{
			ID:         "song1",
			VideoID:    "video1",
			Title:      "bohemian rhapsody",
			Duration:   "5:55",
			URLYoutube: "https://youtube.com/video1",
		},
		{
			ID:         "song2",
			VideoID:    "video2",
			Title:      "stairway to heaven",
			Duration:   "8:02",
			URLYoutube: "https://youtube.com/video2",
		},
		{
			ID:         "song3",
			VideoID:    "video3",
			Title:      "hotel california",
			Duration:   "6:30",
			URLYoutube: "https://youtube.com/video3",
		},
		{
			ID:         "song4",
			VideoID:    "video4",
			Title:      "sweet child o'mine",
			Duration:   "5:56",
			URLYoutube: "https://youtube.com/video4",
		},
		{
			ID:         "song5",
			VideoID:    "video5",
			Title:      "smells like teen spirit",
			Duration:   "4:38",
			URLYoutube: "https://youtube.com/video5",
		},
		{
			ID:         "song6",
			VideoID:    "video6",
			Title:      "love story",
			Duration:   "3:55",
			URLYoutube: "https://youtube.com/video6",
		},
		{
			ID:         "song7",
			VideoID:    "video7",
			Title:      "lover",
			Duration:   "3:41",
			URLYoutube: "https://youtube.com/video7",
		},
	}

	for _, song := range testSongs {
		item, err := attributevalue.MarshalMap(struct {
			entity.Song
			GSIPK string `dynamodbav:"GSI2_PK"`
		}{
			Song:  song,
			GSIPK: "SEARCH#TITLE",
		})
		require.NoError(t, err)

		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      item,
		})
		require.NoError(t, err)
	}

	t.Run("Buscar 'bohemian'", func(t *testing.T) {
		results, err := repo.SearchSongsByTitle(ctx, "Bohemian")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "bohemian rhapsody", results[0].Title)
	})

	t.Run("Buscar 'love'", func(t *testing.T) {
		results, err := repo.SearchSongsByTitle(ctx, "love")
		require.NoError(t, err)
		assert.Len(t, results, 2)
		titles := []string{results[0].Title, results[1].Title}
		assert.Contains(t, titles, "love story")
		assert.Contains(t, titles, "lover")
	})

	t.Run("Buscar 'swee'", func(t *testing.T) {
		results, err := repo.SearchSongsByTitle(ctx, "swee")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "sweet child o'mine", results[0].Title)
	})

	t.Run("Buscar 'z' (sin resultados)", func(t *testing.T) {
		results, err := repo.SearchSongsByTitle(ctx, "z")
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("Buscar 'StAiRwAy' (case-insensitive)", func(t *testing.T) {
		results, err := repo.SearchSongsByTitle(ctx, "StAiRwAy")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "stairway to heaven", results[0].Title)
	})
}
