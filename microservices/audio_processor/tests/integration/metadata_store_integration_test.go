package integration

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/dynamodbservice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestIntegrationMetadataStore(t *testing.T) {
	if testing.Short() {
		t.Skip("Saltando test de integración en modo corto")
	}

	tableName := os.Getenv("DYNAMODB_TABLE_NAME")
	region := os.Getenv("REGION")

	if tableName == "" || region == "" {
		t.Fatal("DYNAMODB_TABLE_NAME y REGION no fueron seteados para los tests de integracion")
	}

	t.Run("SaveAndRetrieveMetadata", func(t *testing.T) {
		// arrange
		store, err := dynamodbservice.NewMetadataStore(tableName, region)
		require.NoError(t, err)

		metadata := model.Metadata{
			ID:         "integration-test-id",
			Title:      "Integration Test Song",
			URLYouTube: "https://www.youtube.com/watch?v=example",
			URLS3:      "https://s3.amazonaws.com/mybucket/integration-test-id",
			Platform:   "YouTube",
			Artist:     "Test Artist",
			Duration:   240,
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
		assert.Equal(t, metadata.Artist, retrievedMetadata.Artist)
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