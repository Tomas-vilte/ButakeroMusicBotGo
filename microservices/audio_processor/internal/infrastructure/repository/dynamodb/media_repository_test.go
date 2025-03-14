//go:build integration

package dynamodb_test

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	dynamodb2 "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/repository/dynamodb"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

func setupDynamoDBContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		Image:        "amazon/dynamodb-local:latest",
		ExposedPorts: []string{"8000/tcp"},
		WaitingFor:   wait.ForLog("Initializing DynamoDB Local"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}
	return container, nil
}

func createDynamoDBTables(ctx context.Context, client *dynamodb.Client) error {
	_, err := client.CreateTable(ctx, &dynamodb.CreateTableInput{
		TableName: aws.String("Songs"),
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
				AttributeName: aws.String("title_lower"),
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
				IndexName: aws.String("TitleLowerIndex"),
				KeySchema: []types.KeySchemaElement{
					{
						AttributeName: aws.String("title_lower"),
						KeyType:       types.KeyTypeHash,
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
	})

	return err
}

func createDynamoDBClient(ctx context.Context, container testcontainers.Container) (*dynamodb.Client, error) {
	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		return nil, err
	}

	cfg, err := awsCfg.LoadDefaultConfig(ctx, awsCfg.WithEndpointResolver(aws.EndpointResolverFunc(
		func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{URL: "http://" + endpoint}, nil
		})))
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(cfg), nil
}

func TestMediaRepositoryDynamoDB_Integration(t *testing.T) {
	ctx := context.Background()

	container, err := setupDynamoDBContainer(ctx)
	assert.NoError(t, err)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	client, err := createDynamoDBClient(ctx, container)
	assert.NoError(t, err)

	err = createDynamoDBTables(ctx, client)
	assert.NoError(t, err)

	log, err := logger.NewDevelopmentLogger()
	assert.NoError(t, err)

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DynamoDB: &config.DynamoDBConfig{
				Tables: config.Tables{
					Songs: "Songs",
				},
			},
		},
	}
	repo, err := dynamodb2.NewMediaRepositoryDynamoDB(cfg, log)
	assert.NoError(t, err)

	media := &model.Media{
		VideoID:    "video123",
		TitleLower: "test song",
		Status:     "processed",
		Message:    "success",
		Metadata: &model.PlatformMetadata{
			Title:        "Test Song",
			DurationMs:   300000,
			URL:          "https://youtube.com/watch?v=video123",
			ThumbnailURL: "https://img.youtube.com/vi/video123/default.jpg",
			Platform:     "YouTube",
		},
		FileData: &model.FileData{
			FilePath: "/path/to/file.mp3",
			FileSize: "10MB",
			FileType: "audio/mpeg",
		},
		ProcessingDate: time.Now(),
		Success:        true,
		Attempts:       1,
		Failures:       0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		PlayCount:      0,
	}

	err = repo.SaveMedia(ctx, media)
	assert.NoError(t, err)

	retrievedMedia, err := repo.GetMedia(ctx, media.VideoID)
	assert.NoError(t, err)
	assert.Equal(t, media.VideoID, retrievedMedia.VideoID)
	assert.Equal(t, media.Status, retrievedMedia.Status)
	assert.Equal(t, media.Message, retrievedMedia.Message)
	assert.Equal(t, media.Metadata.Title, retrievedMedia.Metadata.Title)
	assert.Equal(t, media.Metadata.DurationMs, retrievedMedia.Metadata.DurationMs)
	assert.Equal(t, media.Metadata.URL, retrievedMedia.Metadata.URL)
	assert.Equal(t, media.Metadata.ThumbnailURL, retrievedMedia.Metadata.ThumbnailURL)
	assert.Equal(t, media.Metadata.Platform, retrievedMedia.Metadata.Platform)
	assert.Equal(t, media.FileData.FilePath, retrievedMedia.FileData.FilePath)
	assert.Equal(t, media.FileData.FileSize, retrievedMedia.FileData.FileSize)
	assert.Equal(t, media.FileData.FileType, retrievedMedia.FileData.FileType)
	assert.Equal(t, media.Success, retrievedMedia.Success)
	assert.Equal(t, media.Attempts, retrievedMedia.Attempts)
	assert.Equal(t, media.Failures, retrievedMedia.Failures)
	assert.Equal(t, media.PlayCount, retrievedMedia.PlayCount)

	media.Status = "updated"
	media.Message = "updated message"
	err = repo.UpdateMedia(ctx, media.VideoID, media)
	assert.NoError(t, err)

	updatedMedia, err := repo.GetMedia(ctx, media.VideoID)
	assert.NoError(t, err)
	assert.Equal(t, "updated", updatedMedia.Status)
	assert.Equal(t, "updated message", updatedMedia.Message)

	err = repo.DeleteMedia(ctx, media.VideoID)
	assert.NoError(t, err)

	deletedMedia, err := repo.GetMedia(ctx, media.VideoID)
	assert.Error(t, err)
	assert.Nil(t, deletedMedia)
}

func TestMediaRepositoryDynamoDB_SaveMedia_InvalidID(t *testing.T) {
	ctx := context.Background()

	container, err := setupDynamoDBContainer(ctx)
	assert.NoError(t, err)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	_, err = createDynamoDBClient(ctx, container)
	assert.NoError(t, err)

	log, err := logger.NewDevelopmentLogger()
	assert.NoError(t, err)

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DynamoDB: &config.DynamoDBConfig{
				Tables: config.Tables{
					Songs: "Songs",
				},
			},
		},
	}
	repo, err := dynamodb2.NewMediaRepositoryDynamoDB(cfg, log)
	assert.NoError(t, err)

	media := &model.Media{
		VideoID:    "",
		TitleLower: "test song",
		Status:     "processed",
		Message:    "success",
		Metadata: &model.PlatformMetadata{
			DurationMs:   300000,
			URL:          "https://youtube.com/watch?v=video123",
			ThumbnailURL: "https://img.youtube.com/vi/video123/default.jpg",
			Platform:     "YouTube",
		},
	}

	err = repo.SaveMedia(ctx, media)
	assert.Error(t, err)
	assert.Equal(t, dynamodb2.ErrInvalidVideoID, err)
}

func TestMediaRepositoryDynamoDB_GetMedia_InvalidVideoID(t *testing.T) {
	ctx := context.Background()

	container, err := setupDynamoDBContainer(ctx)
	assert.NoError(t, err)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	_, err = createDynamoDBClient(ctx, container)
	assert.NoError(t, err)

	log, err := logger.NewDevelopmentLogger()
	assert.NoError(t, err)

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DynamoDB: &config.DynamoDBConfig{
				Tables: config.Tables{
					Songs: "Songs",
				},
			},
		},
	}
	repo, err := dynamodb2.NewMediaRepositoryDynamoDB(cfg, log)
	assert.NoError(t, err)

	_, err = repo.GetMedia(ctx, "invalid-video-123")
	assert.Error(t, err)
	assert.Equal(t, dynamodb2.ErrMediaNotFound, err)
}

func TestMediaRepositoryDynamoDB_GetMedia_NotFound(t *testing.T) {
	ctx := context.Background()

	container, err := setupDynamoDBContainer(ctx)
	assert.NoError(t, err)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	_, err = createDynamoDBClient(ctx, container)
	assert.NoError(t, err)

	log, err := logger.NewDevelopmentLogger()
	assert.NoError(t, err)

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DynamoDB: &config.DynamoDBConfig{
				Tables: config.Tables{
					Songs: "Songs",
				},
			},
		},
	}
	repo, err := dynamodb2.NewMediaRepositoryDynamoDB(cfg, log)
	assert.NoError(t, err)

	nonExistentVideoID := "non-existent-video"

	_, err = repo.GetMedia(ctx, nonExistentVideoID)
	assert.Error(t, err)
	assert.Equal(t, dynamodb2.ErrMediaNotFound, err)
}

func TestMediaRepositoryDynamoDB_UpdateMedia_InvalidMetadata(t *testing.T) {
	ctx := context.Background()

	container, err := setupDynamoDBContainer(ctx)
	assert.NoError(t, err)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	_, err = createDynamoDBClient(ctx, container)
	assert.NoError(t, err)

	log, err := logger.NewDevelopmentLogger()
	assert.NoError(t, err)

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DynamoDB: &config.DynamoDBConfig{
				Tables: config.Tables{
					Songs: "Songs",
				},
			},
		},
	}
	repo, err := dynamodb2.NewMediaRepositoryDynamoDB(cfg, log)
	assert.NoError(t, err)

	media := &model.Media{
		VideoID:    "video123",
		Status:     "processed",
		Message:    "success",
		TitleLower: "",
		Metadata: &model.PlatformMetadata{
			DurationMs:   300000,
			URL:          "https://youtube.com/watch?v=video123",
			ThumbnailURL: "https://img.youtube.com/vi/video123/default.jpg",
			Platform:     "",
		},
	}

	err = repo.UpdateMedia(ctx, media.VideoID, media)
	assert.Error(t, err)
	assert.Equal(t, dynamodb2.ErrInvalidMetadata, err)
}

func TestMediaRepositoryDynamoDB_DeleteMedia_NotFound(t *testing.T) {
	ctx := context.Background()

	container, err := setupDynamoDBContainer(ctx)
	assert.NoError(t, err)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	_, err = createDynamoDBClient(ctx, container)
	assert.NoError(t, err)

	log, err := logger.NewDevelopmentLogger()
	assert.NoError(t, err)

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DynamoDB: &config.DynamoDBConfig{
				Tables: config.Tables{
					Songs: "Songs",
				},
			},
		},
	}
	repo, err := dynamodb2.NewMediaRepositoryDynamoDB(cfg, log)
	assert.NoError(t, err)

	nonExistentVideoID := "non-existent-video"

	err = repo.DeleteMedia(ctx, nonExistentVideoID)
	assert.NoError(t, err)
}
