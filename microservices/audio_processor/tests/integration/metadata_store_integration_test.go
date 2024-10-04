package integration

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/persistence/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestIntegrationMetadataStore(t *testing.T) {
	if testing.Short() {
		t.Skip("Saltando test de integración en modo corto")
	}

	cfg := config.Config{
		SongsTable: os.Getenv("DYNAMODB_TABLE_NAME_SONGS"),
		Region:     os.Getenv("REGION"),
		AccessKey:  os.Getenv("ACCESS_KEY"),
		SecretKey:  os.Getenv("SECRET_KEY"),
	}

	if cfg.SongsTable == "" || cfg.Region == "" {
		t.Fatal("DYNAMODB_TABLE_NAME_SONGS y REGION no fueron seteados para los tests de integracion")
	}

	t.Run("SaveAndRetrieveMetadata", func(t *testing.T) {
		// arrange
		store, err := dynamodb.NewMetadataStore(cfg)
		require.NoError(t, err)

		metadata := model.Metadata{
			ID:         "integration-test-id",
			Title:      "Integration Test Song",
			URLYouTube: "https://www.youtube.com/watch?v=example",
			URLS3:      "https://s3.amazonaws.com/mybucket/integration-test-id",
			Platform:   "YouTube",
			Duration:   "240",
		}

		// act SaveMetadata
		err = store.SaveMetadata(context.Background(), metadata)
		require.NoError(t, err)

		retrievedMetadata, err := store.GetMetadata(context.Background(), metadata.ID)
		require.NoError(t, err)

		// assert - chequeamos si la metadata se guardo correctamente
		assert.Equal(t, metadata.ID, retrievedMetadata.ID)
		assert.Equal(t, metadata.Title, retrievedMetadata.Title)
		assert.Equal(t, metadata.URLYouTube, retrievedMetadata.URLYouTube)
		assert.Equal(t, metadata.URLS3, retrievedMetadata.URLS3)
		assert.Equal(t, metadata.Platform, retrievedMetadata.Platform)
		assert.Equal(t, metadata.Duration, retrievedMetadata.Duration)

		// act - delete metadata
		err = store.DeleteMetadata(context.Background(), metadata.ID)
		require.NoError(t, err)

		// act -  intentamos recuperar los datos eliminados
		_, err = store.GetMetadata(context.Background(), metadata.ID)

		// assert no se deben encontrar datos
		assert.Error(t, err)
	})
}
