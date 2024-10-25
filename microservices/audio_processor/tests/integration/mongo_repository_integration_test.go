package integration

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	mongoRepo "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/repository/mongodb"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"log"
	"os"
	"testing"
)

var mongoContainer *mongodb.MongoDBContainer

func TestMain(t *testing.M) {
	var err error
	mongoContainer, err = mongodb.Run(context.Background(), "mongodb:6")
	if err != nil {
		log.Fatalf("Error al iniciar el contendor de MongoDB: %v", err)
	}

	code := t.Run()

	defer func() {
		if err := mongoContainer.Terminate(context.Background()); err != nil {
			log.Fatalf("Error al terminar el contendor de MongoDB: %v", err)
		}
	}()

	os.Exit(code)
}

func setupMongoRepository(t *testing.T) (*mongoRepo.MongoMetadataRepository, func()) {
	ctx := context.Background()
	logging, err := logger.NewZapLogger()
	if err != nil {
		logging.Error("Error al crear el logger", zap.Error(err))
	}
	endpoint, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		logging.Error("Error al obtener la conexion:", zap.Error(err))
	}
	clientOpions := options.Client().ApplyURI(endpoint)

	client, err := mongo.Connect(ctx, clientOpions)
	assert.NoError(t, err)

	collection := client.Database("test_db").Collection("metadata")
	repo := mongoRepo.NewMongoMetadataRepository(mongoRepo.MongoMetadataOptions{
		Collection: collection,
		Log:        logging,
	})

	cleanup := func() {
		err := client.Disconnect(context.Background())
		assert.NoError(t, err)
	}

	return repo, cleanup
}

func TestMongoMetadataRepository_Integration(t *testing.T) {
	repo, cleanup := setupMongoRepository(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("SaveMetadata", func(t *testing.T) {
		// arrange
		metadata := &model.Metadata{
			Title:      "Test Song",
			URLYouTube: "https://youtube.com/test",
			Thumbnail:  "https://img.youtube.com/test",
			Platform:   "youtube",
			Duration:   "3:45",
		}

		// act
		err := repo.SaveMetadata(ctx, metadata)
		assert.NoError(t, err)
		assert.NotEmpty(t, metadata.ID, "El ID no debería estar vacío después de guardar")

		assert.NoError(t, err)
		assert.NotEmpty(t, metadata.ID)

		// cleanup
		_ = repo.DeleteMetadata(context.Background(), metadata.ID)
	})

	t.Run("GetMetadata", func(t *testing.T) {
		// arrange
		metadata := &model.Metadata{
			Title:      "Test Song 2",
			URLYouTube: "https://youtube.com/test2",
			Thumbnail:  "https://img.youtube.com/test2",
			Platform:   "youtube",
			Duration:   "4:20",
		}

		err := repo.SaveMetadata(context.Background(), metadata)
		assert.NoError(t, err)

		// act
		retrieved, err := repo.GetMetadata(ctx, metadata.ID)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, metadata.Title, retrieved.Title)
		assert.Equal(t, metadata.URLYouTube, retrieved.URLYouTube)
		assert.Equal(t, metadata.Platform, retrieved.Platform)

		// cleanup
		if metadata.ID != "" {
			err = repo.DeleteMetadata(ctx, metadata.ID)
			require.NoError(t, err)
		}
	})

	t.Run("DeleteMetadata", func(t *testing.T) {
		// arrange
		metadata := &model.Metadata{
			Title:      "Test Song 3",
			URLYouTube: "https://youtube.com/test3",
			Thumbnail:  "https://img.youtube.com/test3",
			Platform:   "youtube",
			Duration:   "2:30",
		}

		err := repo.SaveMetadata(context.Background(), metadata)
		assert.NoError(t, err)
		require.NotEmpty(t, metadata.ID, "El ID debe estar presente después de guardar")

		// act
		err = repo.DeleteMetadata(context.Background(), metadata.ID)

		// assert
		assert.NoError(t, err)

		// verify deletion
		_, err = repo.GetMetadata(ctx, metadata.ID)
	})

	t.Run("GetMetadata Not Found", func(t *testing.T) {
		// arrange
		nonExistingID := "non-existing-id"

		// act
		_, err := repo.GetMetadata(context.Background(), nonExistingID)

		// assert
		assert.ErrorIs(t, err, mongoRepo.ErrMetadataNotFound)
	})

	t.Run("SaveMetadata Invalid Data", func(t *testing.T) {
		// arrange
		invalidMetadata := &model.Metadata{
			URLYouTube: "https://youtube.com/test",
			Platform:   "YouTube",
		}

		// act
		err := repo.SaveMetadata(context.Background(), invalidMetadata)

		assert.ErrorIs(t, err, mongoRepo.ErrInvalidMetadata)
	})

	t.Run("SaveMetadata Duplicate ID", func(t *testing.T) {
		// arrange
		metadata := model.Metadata{
			ID:         "test-duplicate-id",
			Title:      "Test Song",
			URLYouTube: "https://youtube.com/test",
			Thumbnail:  "https://img.youtube.com/test",
			Platform:   "youtube",
			Duration:   "3:45",
		}

		// act - Intentar guardar el mismo ID 2 veces
		err := repo.SaveMetadata(context.Background(), &metadata)
		assert.NoError(t, err)
		err = repo.SaveMetadata(context.Background(), &metadata)

		// assert
		assert.Error(t, err)

		// cleanup
		if metadata.ID != "" {
			err = repo.DeleteMetadata(context.Background(), metadata.ID)
			assert.NoError(t, err)
		}
	})

	t.Run("GetMetadata Empty ID", func(t *testing.T) {
		_, err := repo.GetMetadata(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID no puede estar vacio")
	})

	t.Run("DeleteMetadata Empty ID", func(t *testing.T) {
		err := repo.DeleteMetadata(context.Background(), "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID no puede estar vacio")
	})

}
