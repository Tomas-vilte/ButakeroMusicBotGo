package integration

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/dynamodbservice"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestIntegrationOperationStore(t *testing.T) {
	if testing.Short() {
		t.Skip("Saltando test de integración en modo corto")
	}

	tableName := os.Getenv("DYNAMODB_TABLE_NAME_OPERATION")
	region := os.Getenv("REGION")

	if tableName == "" || region == "" {
		t.Fatal("DYNAMODB_TABLE_NAME_OPERATION y REGION no fueron seteados para los tests de integracion")
	}

	store, err := dynamodbservice.NewOperationStore(tableName, region)
	require.NoError(t, err)

	t.Run("SaveAndRetrieveOperationResult", func(t *testing.T) {
		// arrange
		result := createTestOperationResult("integration-test-id-1")

		// act SaveOperationResult
		err = store.SaveOperationResult(context.Background(), result)
		require.NoError(t, err)

		// act GetOperationResult
		retrievedResult, err := store.GetOperationResult(context.Background(), result.ID, result.SongID)
		require.NoError(t, err)

		// assert
		assertOperationResultEqual(t, result, *retrievedResult)

		// cleanup
		err = store.DeleteOperationResult(context.Background(), result.ID, result.SongID)
		require.NoError(t, err)
	})

	t.Run("UpdateOperationResult", func(t *testing.T) {
		// arrange
		result := createTestOperationResult("integration-test-id-2")

		// act SaveOperationResult
		err = store.SaveOperationResult(context.Background(), result)
		require.NoError(t, err)

		// act update operation result
		result.Status = "completed"
		result.Message = "Operation completed successfully"
		result.Data = "Updated data"
		err = store.SaveOperationResult(context.Background(), result)
		require.NoError(t, err)

		// act GetOperationResult
		updatedResult, err := store.GetOperationResult(context.Background(), result.ID, result.SongID)
		require.NoError(t, err)

		// assert
		assert.Equal(t, "completed", updatedResult.Status)
		assert.Equal(t, "Operation completed successfully", updatedResult.Message)
		assert.Equal(t, "Updated data", updatedResult.Data)

		// cleanup
		err = store.DeleteOperationResult(context.Background(), result.ID, result.SongID)
		require.NoError(t, err)
	})

	t.Run("DeleteOperationResult", func(t *testing.T) {
		// arrange
		result := createTestOperationResult("integration-test-id-3")

		// act SaveOperationResult
		err = store.SaveOperationResult(context.Background(), result)
		require.NoError(t, err)

		// act DeleteOperationResult
		err = store.DeleteOperationResult(context.Background(), result.ID, result.SongID)
		require.NoError(t, err)

		// act Intentar obtener el resultado de la operación eliminada
		_, err = store.GetOperationResult(context.Background(), result.ID, result.SongID)

		// assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resultado de operación no encontrado")
	})

	t.Run("GetNonExistentOperationResult", func(t *testing.T) {
		// act
		_, err := store.GetOperationResult(context.Background(), "non-existent-id", "non-existent-song-id")

		// assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resultado de operación no encontrado")
	})
}

func createTestOperationResult(songID string) model.OperationResult {
	return model.OperationResult{
		ID:      uuid.New().String(),
		SongID:  songID,
		Status:  "in_progress",
		Message: "Operation is in progress",
		Data:    "Some initial data",
	}
}

func assertOperationResultEqual(t *testing.T, expected, actual model.OperationResult) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.SongID, actual.SongID)
	assert.Equal(t, expected.Status, actual.Status)
	assert.Equal(t, expected.Message, actual.Message)
	assert.Equal(t, expected.Data, actual.Data)
}