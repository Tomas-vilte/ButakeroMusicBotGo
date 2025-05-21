//go:build integration

package mongodb_test

import (
	"context"
	errorsApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/repository/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
)

func setupMongoDB(t *testing.T) (*mongo.Client, func()) {
	ctx := context.Background()

	container, err := mongodb.NewMongoDBContainer(ctx, mongodb.DefaultMongoDBConfig())
	require.NoError(t, err, "Error al crear el contenedor de MongoDB")

	err = container.Connect(ctx)
	require.NoError(t, err, "Error al conectar a MongoDB")

	return container.Client, func() {
		err := container.Cleanup(ctx)
		require.NoError(t, err, "Error al limpiar el contenedor de MongoDB")
	}
}

func TestMediaRepository_SaveMedia(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	ctx := context.Background()
	log, err := logger.NewDevelopmentLogger()
	require.NoError(t, err, "Error al crear el logger")

	collection := client.Database("test_db").Collection("songs")
	repo, err := mongodb.NewMediaRepository(mongodb.MediaRepositoryOptions{
		Collection: collection,
		Log:        log,
	})
	require.NoError(t, err, "Error al crear el repositorio")

	media := &model.Media{
		VideoID:    "video123",
		Status:     "processed",
		Message:    "success",
		TitleLower: "test song",
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
	assert.NoError(t, err, "Error al guardar el registro de media")

	retrievedMedia, err := repo.GetMediaByID(ctx, media.VideoID)
	assert.NoError(t, err, "Error al obtener el registro de media")
	assert.Equal(t, media.VideoID, retrievedMedia.VideoID)
	assert.Equal(t, media.Status, retrievedMedia.Status)
	assert.Equal(t, media.Message, retrievedMedia.Message)
	assert.Equal(t, media.Metadata.Title, retrievedMedia.Metadata.Title)
	assert.Equal(t, media.TitleLower, retrievedMedia.TitleLower)
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
}

func TestMediaRepository_GetMedia_NotFound(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	ctx := context.Background()
	log, err := logger.NewDevelopmentLogger()
	require.NoError(t, err, "Error al crear el logger")

	collection := client.Database("test_db").Collection("songs")
	repo, err := mongodb.NewMediaRepository(mongodb.MediaRepositoryOptions{
		Collection: collection,
		Log:        log,
	})
	require.NoError(t, err, "Error al crear el repositorio")

	nonExistentVideoID := "non-existent-video"

	_, err = repo.GetMediaByID(ctx, nonExistentVideoID)
	assert.Error(t, err, "Se esperaba un error al obtener un registro inexistente")
	assert.Equal(t, errorsApp.ErrCodeMediaNotFound, err, "El error no es el esperado")
}

func TestMediaRepository_UpdateMedia(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	ctx := context.Background()
	log, err := logger.NewDevelopmentLogger()
	require.NoError(t, err, "Error al crear el logger")

	collection := client.Database("test_db").Collection("songs")
	repo, err := mongodb.NewMediaRepository(mongodb.MediaRepositoryOptions{
		Collection: collection,
		Log:        log,
	})
	require.NoError(t, err, "Error al crear el repositorio")

	media := &model.Media{
		VideoID:    "video123",
		Status:     "processed",
		Message:    "success",
		TitleLower: "test song",
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
	assert.NoError(t, err, "Error al guardar el registro de media")

	media.Status = "updated"
	media.Message = "updated message"
	err = repo.UpdateMedia(ctx, media.VideoID, media)
	assert.NoError(t, err, "Error al actualizar el registro de media")

	updatedMedia, err := repo.GetMediaByID(ctx, media.VideoID)
	assert.NoError(t, err, "Error al obtener el registro de media actualizado")
	assert.Equal(t, "updated", updatedMedia.Status)
	assert.Equal(t, "updated message", updatedMedia.Message)
}

func TestMediaRepository_GetMediaByTitle(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	ctx := context.Background()
	log, err := logger.NewDevelopmentLogger()
	require.NoError(t, err, "Error al crear el logger")

	collection := client.Database("test_db").Collection("songs")
	repo, err := mongodb.NewMediaRepository(mongodb.MediaRepositoryOptions{
		Collection: collection,
		Log:        log,
	})
	require.NoError(t, err, "Error al crear el repositorio")

	t.Run("SearchSongsByTitle", func(t *testing.T) {
		testSongs := []*model.Media{
			{
				VideoID:    "video123",
				TitleLower: "test song one",
				Metadata: &model.PlatformMetadata{
					Title:      "Test Song One",
					DurationMs: 3245,
					URL:        "https://youtube.com/video123",
				},
			},
			{
				VideoID:    "video456",
				TitleLower: "another test song",
				Metadata: &model.PlatformMetadata{
					Title:      "Another Test Song",
					DurationMs: 2323,
					URL:        "https://youtube.com/video456",
				},
			},
			{
				VideoID:    "video789",
				TitleLower: "something completely different",
				Metadata: &model.PlatformMetadata{
					Title:      "Something Completely Different",
					DurationMs: 4242,
					URL:        "https://youtube.com/video789",
				},
			},
		}

		for _, song := range testSongs {
			_, err := collection.InsertOne(ctx, song)
			assert.NoError(t, err)
		}

		defer func() {
			_, err := collection.DeleteMany(ctx, bson.M{
				"_id": bson.M{"$in": []string{"video123", "video456", "video789"}},
			})
			assert.NoError(t, err)
		}()

		// Act
		results, err := repo.GetMediaByTitle(ctx, "Test")

		// Assert
		assert.NoError(t, err)
		assert.Len(t, results, 2)
		assert.Contains(t, []string{results[0].Metadata.Title, results[1].Metadata.Title}, "Test Song One")
		assert.Contains(t, []string{results[0].Metadata.Title, results[1].Metadata.Title}, "Another Test Song")
	})
}

func TestMediaRepository_DeleteMedia(t *testing.T) {
	client, cleanup := setupMongoDB(t)
	defer cleanup()

	ctx := context.Background()
	log, err := logger.NewDevelopmentLogger()
	require.NoError(t, err, "Error al crear el logger")

	collection := client.Database("test_db").Collection("songs")
	repo, err := mongodb.NewMediaRepository(mongodb.MediaRepositoryOptions{
		Collection: collection,
		Log:        log,
	})
	require.NoError(t, err, "Error al crear el repositorio")

	media := &model.Media{
		VideoID:    "video123",
		Status:     "processed",
		Message:    "success",
		TitleLower: "test song",
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
	assert.NoError(t, err, "Error al guardar el registro de media")

	err = repo.DeleteMedia(ctx, media.VideoID)
	assert.NoError(t, err, "Error al eliminar el registro de media")

	_, err = repo.GetMediaByID(ctx, media.VideoID)
	assert.Error(t, err, "Se esperaba un error al obtener un registro eliminado")
	assert.Equal(t, errorsApp.ErrCodeMediaNotFound, err, "El error no es el esperado")
}
