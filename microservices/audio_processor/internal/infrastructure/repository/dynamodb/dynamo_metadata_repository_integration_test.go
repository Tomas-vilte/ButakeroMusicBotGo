//go:build integration

package dynamodb

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestIntegrationMetadataStore(t *testing.T) {
	if testing.Short() {
		t.Skip("Saltando test de integración en modo corto")
	}

	log, err := logger.NewDevelopmentLogger()
	require.NoError(t, err)

	cfg := &config.Config{
		AWS: config.AWSConfig{
			Region: os.Getenv("AWS_REGION"),
		},
		Database: config.DatabaseConfig{
			DynamoDB: &config.DynamoDBConfig{
				Tables: config.Tables{
					Songs: os.Getenv("DYNAMODB_TABLE_NAME_SONGS"),
				},
			},
		},
	}

	if cfg.Database.DynamoDB.Tables.Songs == "" || cfg.AWS.Region == "" {
		t.Fatal("DYNAMODB_TABLE_NAME_SONGS y REGION no fueron seteados para los tests de integracion")
	}

	t.Run("SaveAndRetrieveMetadata", func(t *testing.T) {
		// arrange
		store, err := NewMetadataStore(cfg, log)
		require.NoError(t, err)

		metadata := &model.Metadata{
			ID:       "integration-test-id",
			Title:    "Integration Test Song",
			URL:      "https://www.youtube.com/watch?v=example",
			Platform: "YouTube",
			Duration: "240",
		}

		// act SaveMetadata
		err = store.SaveMetadata(context.Background(), metadata)
		require.NoError(t, err)

		retrievedMetadata, err := store.GetMetadata(context.Background(), metadata.ID)
		require.NoError(t, err)

		// assert - chequeamos si la metadata se guardo correctamente
		assert.Equal(t, metadata.ID, retrievedMetadata.ID)
		assert.Equal(t, metadata.Title, retrievedMetadata.Title)
		assert.Equal(t, metadata.URL, retrievedMetadata.URL)
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
