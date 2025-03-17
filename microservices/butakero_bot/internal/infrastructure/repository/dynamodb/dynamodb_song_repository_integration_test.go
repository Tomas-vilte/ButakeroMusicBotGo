//go:build integration

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
	"strings"
	"testing"
	"time"
)

const (
	tableName   = "test_songs"
	defaultPort = "8000"
)

func createTestTableWithIndexes(ctx context.Context, client *dynamodb.Client) error {
	input := &dynamodb.CreateTableInput{
		TableName: aws.String(tableName),
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: aws.String("PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI1PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: aws.String("GSI1SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: aws.String("PK"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: aws.String("SK"),
				KeyType:       types.KeyTypeRange,
			},
		},
		GlobalSecondaryIndexes: []types.GlobalSecondaryIndex{
			{
				IndexName: aws.String("GSI1"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("GSI1PK"),
						KeyType:       types.KeyTypeHash,
					},
					{
						AttributeName: aws.String("GSI1SK"),
						KeyType:       types.KeyTypeRange,
					},
				},
				Projection: &types.Projection{
					ProjectionType: types.ProjectionTypeAll,
				},
				ProvisionedThroughput: &types.ProvisionedThroughput{
					ReadCapacityUnits:  aws.Int64(5),
					WriteCapacityUnits: aws.Int64(5),
				},
			},
		},
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  aws.Int64(5),
			WriteCapacityUnits: aws.Int64(5),
		},
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
	logger, err := logging.NewDevelopmentLogger()
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

	testSongs := []entity.SongEntity{
		{
			PK:         "6swmTBVI83k",
			SK:         "METADATA",
			TitleLower: "lil nas x - montero (call me by your name) (official video)",
			Metadata: entity.Metadata{
				Title:      "Lil Nas X - MONTERO (Call Me By Your Name) (Official Video)",
				DurationMs: 190000,
				URL:        "https://youtube.com/watch?v=6swmTBVI83k",
			},
		},
		{
			PK:         "dQw4w9WgXcQ",
			SK:         "METADATA",
			TitleLower: "rick astley - never gonna give you up (official music video)",
			Metadata: entity.Metadata{
				Title:      "Rick Astley - Never Gonna Give You Up (Official Music Video)",
				DurationMs: 213000,
				URL:        "https://youtube.com/watch?v=dQw4w9WgXcQ",
			},
		},
		{
			PK:         "kJQP7kiw5Fk",
			SK:         "METADATA",
			TitleLower: "luis fonsi - despacito ft. daddy yankee",
			Metadata: entity.Metadata{
				Title:      "Luis Fonsi - Despacito ft. Daddy Yankee",
				DurationMs: 229000,
				URL:        "https://youtube.com/watch?v=kJQP7kiw5Fk",
			},
		},
	}

	for _, song := range testSongs {
		titleLower := strings.ToLower(song.Metadata.Title)
		song.TitleLower = titleLower
		song.PK = fmt.Sprintf("VIDEO#" + song.PK)
		song.GSI1PK = "SONG"
		song.GSI1SK = titleLower
		item, err := attributevalue.MarshalMap(song)
		require.NoError(t, err)

		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      item,
		})
		require.NoError(t, err)
	}

	// Casos de prueba
	t.Run("Buscar 'montero' (búsqueda parcial)", func(t *testing.T) {
		results, err := repo.SearchSongsByTitle(ctx, "lil nas x - montero")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "lil nas x - montero (call me by your name) (official video)", results[0].TitleLower)
	})

	t.Run("Buscar 'never gonna' (búsqueda parcial)", func(t *testing.T) {
		results, err := repo.SearchSongsByTitle(ctx, "rick astley - never gonna")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "rick astley - never gonna give you up (official music video)", results[0].TitleLower)
	})

	t.Run("Buscar 'despacito' (búsqueda exacta)", func(t *testing.T) {
		results, err := repo.SearchSongsByTitle(ctx, "luis fonsi - despacito")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "luis fonsi - despacito ft. daddy yankee", results[0].TitleLower)
	})

	t.Run("Buscar 'z' (sin resultados)", func(t *testing.T) {
		results, err := repo.SearchSongsByTitle(ctx, "z")
		require.NoError(t, err)
		assert.Empty(t, results)
	})
}
