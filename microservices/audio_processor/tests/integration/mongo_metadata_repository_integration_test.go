package integration

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/repository/mongodb"
	mongoHelper "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/tests/testutil/mongodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMongoMetadataRepository_Integration(t *testing.T) {
	helper := mongoHelper.NewTestHelper(t)
	defer helper.Cleanup(t)

	repo, err := mongodb.NewMongoMetadataRepository(mongodb.MongoMetadataOptions{
		Collection: helper.MongoDB.GetCollection("metadata"),
		Log:        helper.Logger,
	})
	assert.NoError(t, err)

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
		err := repo.SaveMetadata(helper.Context, metadata)
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
		retrieved, err := repo.GetMetadata(helper.Context, metadata.ID)

		// assert
		assert.NoError(t, err)
		assert.Equal(t, metadata.Title, retrieved.Title)
		assert.Equal(t, metadata.URLYouTube, retrieved.URLYouTube)
		assert.Equal(t, metadata.Platform, retrieved.Platform)

		// cleanup
		if metadata.ID != "" {
			err = repo.DeleteMetadata(helper.Context, metadata.ID)
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
		_, err = repo.GetMetadata(helper.Context, metadata.ID)
	})

	t.Run("GetMetadata Not Found", func(t *testing.T) {
		// arrange
		nonExistingID := "non-existing-id"

		// act
		_, err := repo.GetMetadata(context.Background(), nonExistingID)

		// assert
		assert.ErrorIs(t, err, mongodb.ErrMetadataNotFound)
	})

	t.Run("SaveMetadata Invalid Data", func(t *testing.T) {
		// arrange
		invalidMetadata := &model.Metadata{
			URLYouTube: "https://youtube.com/test",
			Platform:   "YouTube",
		}

		// act
		err := repo.SaveMetadata(context.Background(), invalidMetadata)

		assert.ErrorIs(t, err, mongodb.ErrInvalidMetadata)
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
		_, err := repo.GetMetadata(helper.Context, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID no puede estar vacio")
	})

	t.Run("DeleteMetadata Empty ID", func(t *testing.T) {
		err := repo.DeleteMetadata(context.Background(), "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "ID no puede estar vacio")
	})

}
