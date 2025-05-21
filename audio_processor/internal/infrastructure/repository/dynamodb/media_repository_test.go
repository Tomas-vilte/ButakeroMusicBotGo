//go:build integration

package dynamodb_test

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	errorsApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
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

const (
	tableName = "test_songs"
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

func createDynamoDBClient(ctx context.Context, container testcontainers.Container) (*dynamodb.Client, error) {
	endpoint, err := container.Endpoint(ctx, "")
	if err != nil {
		return nil, err
	}

	cfg, err := awsCfg.LoadDefaultConfig(ctx, awsCfg.WithRegion("us-east-1"), awsCfg.WithEndpointResolver(
		aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			return aws.Endpoint{URL: "http://" + endpoint}, nil
		})))
	if err != nil {
		return nil, err
	}

	return dynamodb.NewFromConfig(cfg), nil
}

func setupTestRepository(ctx context.Context, client *dynamodb.Client) (*dynamodb2.MediaRepositoryDynamoDB, error) {
	log, err := logger.NewDevelopmentLogger()
	if err != nil {
		return nil, err
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			DynamoDB: &config.DynamoDBConfig{
				Tables: config.Tables{
					Songs: tableName,
				},
			},
		},
	}

	repo := dynamodb2.NewMediaRepositoryDynamoDB(cfg, log, dynamodb2.WithClient(client))

	return repo, nil
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

	err = createTestTableWithIndexes(ctx, client)
	assert.NoError(t, err)

	repo, err := setupTestRepository(ctx, client)
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

	retrievedMedia, err := repo.GetMediaByID(ctx, media.VideoID)
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

	updatedMedia, err := repo.GetMediaByID(ctx, media.VideoID)
	assert.NoError(t, err)
	assert.Equal(t, "updated", updatedMedia.Status)
	assert.Equal(t, "updated message", updatedMedia.Message)

	err = repo.DeleteMedia(ctx, media.VideoID)
	assert.NoError(t, err)

	deletedMedia, err := repo.GetMediaByID(ctx, media.VideoID)
	assert.Error(t, err)
	assert.Nil(t, deletedMedia)
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

	client, err := createDynamoDBClient(ctx, container)
	assert.NoError(t, err)

	err = createTestTableWithIndexes(ctx, client)
	assert.NoError(t, err)

	repo, err := setupTestRepository(ctx, client)
	assert.NoError(t, err)

	_, err = repo.GetMediaByID(ctx, "invalid-video-123")
	assert.Error(t, err)
	assert.Equal(t, errorsApp.ErrCodeMediaNotFound, err)
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

	client, err := createDynamoDBClient(ctx, container)
	assert.NoError(t, err)

	err = createTestTableWithIndexes(ctx, client)
	assert.NoError(t, err)

	repo, err := setupTestRepository(ctx, client)
	assert.NoError(t, err)

	nonExistentVideoID := "non-existent-video"

	_, err = repo.GetMediaByID(ctx, nonExistentVideoID)
	assert.Error(t, err)
	assert.Equal(t, errorsApp.ErrCodeMediaNotFound, err)
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

	client, err := createDynamoDBClient(ctx, container)
	assert.NoError(t, err)

	err = createTestTableWithIndexes(ctx, client)
	assert.NoError(t, err)

	repo, err := setupTestRepository(ctx, client)
	assert.NoError(t, err)

	nonExistentVideoID := "non-existent-video"

	err = repo.DeleteMedia(ctx, nonExistentVideoID)
	assert.NoError(t, err)
}

func TestMediaRepositoryDynamoDB_GetMediaByTitle(t *testing.T) {
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

	err = createTestTableWithIndexes(ctx, client)
	assert.NoError(t, err)

	repo, err := setupTestRepository(ctx, client)
	assert.NoError(t, err)

	songs := []*model.Media{
		{
			VideoID:    "video1",
			TitleLower: "test song one",
			Status:     "processed",
			Message:    "success",
			Metadata: &model.PlatformMetadata{
				Title:        "Test Song One",
				DurationMs:   300000,
				URL:          "https://youtube.com/watch?v=video1",
				ThumbnailURL: "https://img.youtube.com/vi/video1/default.jpg",
				Platform:     "YouTube",
			},
			FileData: &model.FileData{
				FilePath: "/path/to/file1.mp3",
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
		},
		{
			VideoID:    "video2",
			TitleLower: "test song two",
			Status:     "processed",
			Message:    "success",
			Metadata: &model.PlatformMetadata{
				Title:        "Test Song Two",
				DurationMs:   240000,
				URL:          "https://youtube.com/watch?v=video2",
				ThumbnailURL: "https://img.youtube.com/vi/video2/default.jpg",
				Platform:     "YouTube",
			},
			FileData: &model.FileData{
				FilePath: "/path/to/file2.mp3",
				FileSize: "8MB",
				FileType: "audio/mpeg",
			},
			ProcessingDate: time.Now(),
			Success:        true,
			Attempts:       1,
			Failures:       0,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			PlayCount:      0,
		},
		{
			VideoID:    "video3",
			TitleLower: "another music",
			Status:     "processed",
			Message:    "success",
			Metadata: &model.PlatformMetadata{
				Title:        "Another Music",
				DurationMs:   180000,
				URL:          "https://youtube.com/watch?v=video3",
				ThumbnailURL: "https://img.youtube.com/vi/video3/default.jpg",
				Platform:     "YouTube",
			},
			FileData: &model.FileData{
				FilePath: "/path/to/file3.mp3",
				FileSize: "6MB",
				FileType: "audio/mpeg",
			},
			ProcessingDate: time.Now(),
			Success:        true,
			Attempts:       1,
			Failures:       0,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			PlayCount:      0,
		},
	}

	for _, song := range songs {
		err = repo.SaveMedia(ctx, song)
		assert.NoError(t, err)
	}

	t.Run("Exact match", func(t *testing.T) {
		results, err := repo.GetMediaByTitle(ctx, "test song one")
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "VIDEO#video1", results[0].PK)
		assert.Equal(t, "Test Song One", results[0].Metadata.Title)
	})

	t.Run("Prefix match", func(t *testing.T) {
		results, err := repo.GetMediaByTitle(ctx, "test song")
		assert.NoError(t, err)
		assert.Len(t, results, 2)

		videoIDs := []string{results[0].PK, results[1].PK}
		assert.Contains(t, videoIDs, "VIDEO#video1")
		assert.Contains(t, videoIDs, "VIDEO#video2")
	})

	t.Run("Case insensitive", func(t *testing.T) {
		results, err := repo.GetMediaByTitle(ctx, "TEST SoNg OnE")
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "VIDEO#video1", results[0].PK)
	})

	t.Run("No results", func(t *testing.T) {
		results, err := repo.GetMediaByTitle(ctx, "nonexistent song")
		assert.NoError(t, err)
		assert.Len(t, results, 0)
	})

	t.Run("Partial word match", func(t *testing.T) {
		results, err := repo.GetMediaByTitle(ctx, "anoth")
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "VIDEO#video3", results[0].PK)
	})
}
