package integration

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/repository/mongodb"
	mongoHelper "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/tests/testutil/mongodb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestOperationRepository(t *testing.T) {
	helper := mongoHelper.NewTestHelper(t)
	defer helper.Cleanup(t)

	repo, err := mongodb.NewOperationRepository(mongodb.OperationOptions{
		Collection: helper.MongoDB.GetCollection("operations"),
		Log:        helper.Logger,
	})
	assert.NoError(t, err)

	t.Run("NewOperationRepository", func(t *testing.T) {
		t.Run("should create repository successfully", func(t *testing.T) {
			// act
			repo, err := mongodb.NewOperationRepository(mongodb.OperationOptions{
				Collection: helper.MongoDB.GetCollection("operations"),
				Log:        helper.Logger,
			})

			assert.NoError(t, err)
			assert.NotNil(t, repo)
		})

		t.Run("should return error when collection is nil", func(t *testing.T) {
			// act
			repo, err := mongodb.NewOperationRepository(mongodb.OperationOptions{
				Collection: nil,
				Log:        helper.Logger,
			})

			// assert
			assert.Error(t, err)
			assert.Nil(t, repo)
			assert.Contains(t, err.Error(), "collection no puede ser nil")
		})

		t.Run("should return error when logger is nil", func(t *testing.T) {
			// act
			repo, err := mongodb.NewOperationRepository(mongodb.OperationOptions{
				Collection: helper.MongoDB.GetCollection("operations"),
				Log:        nil,
			})

			// assert
			assert.Error(t, err)
			assert.Nil(t, repo)
			assert.Contains(t, err.Error(), "logger no puede ser nil")
		})
	})

	t.Run("SaveOperationResult", func(t *testing.T) {
		t.Run("should save operation successfully", func(t *testing.T) {
			// arrange
			operation := &model.OperationResult{
				SK:     "test-song-1",
				Status: "pending",
			}

			// act
			err := repo.SaveOperationsResult(helper.Context, operation)

			// assert
			assert.NoError(t, err)
			assert.NotEmpty(t, operation.PK)

			// verificar si se guardo correctamente
			saved, err := repo.GetOperationResult(helper.Context, operation.PK, operation.SK)
			assert.NoError(t, err)
			assert.Equal(t, operation.Status, saved.Status)
			assert.Equal(t, operation.SK, saved.SK)
		})

		t.Run("should return error when operation is nil", func(t *testing.T) {
			// act
			err := repo.SaveOperationsResult(helper.Context, nil)

			// assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "result no puede ser nil")
		})

		t.Run("should generate PK when not provided", func(t *testing.T) {
			// arrange
			operation := &model.OperationResult{
				SK:     "test-song-2",
				Status: "pending",
			}

			// act
			err := repo.SaveOperationsResult(helper.Context, operation)

			// assert
			assert.NoError(t, err)
			assert.NotEmpty(t, operation.PK)
		})
	})

	t.Run("GetOperationResult", func(t *testing.T) {
		t.Run("should get existing operation successfully", func(t *testing.T) {
			// arrange
			operation := &model.OperationResult{
				SK:     "test-song-3",
				Status: "pending",
			}

			err := repo.SaveOperationsResult(helper.Context, operation)
			assert.NoError(t, err)

			// act
			result, err := repo.GetOperationResult(helper.Context, operation.PK, operation.SK)

			// assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, operation.Status, result.Status)
			assert.Equal(t, operation.SK, result.SK)
		})

		t.Run("should return ErrOperationNotFound for non-existent operation", func(t *testing.T) {
			// Act
			result, err := repo.GetOperationResult(helper.Context, "non-existent", "test-song")

			// Assert
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.ErrorIs(t, err, mongodb.ErrOperationNotFound)
		})

		t.Run("should return error with empty parameters", func(t *testing.T) {
			// act
			result, err := repo.GetOperationResult(helper.Context, "", "")

			// assert
			assert.Error(t, err)
			assert.Nil(t, result)
			assert.Contains(t, err.Error(), "id y songID son requeridos")
		})
	})

	t.Run("UpdateOperationStatus", func(t *testing.T) {
		t.Run("should update status successfully", func(t *testing.T) {
			// arrange
			operation := &model.OperationResult{
				SK:     "test-song-4",
				Status: "pending",
			}

			err := repo.SaveOperationsResult(helper.Context, operation)
			assert.NoError(t, err)

			// Act
			err = repo.UpdateOperationStatus(helper.Context, operation.PK, operation.SK, "completed")

			// Assert
			assert.NoError(t, err)

			// Verificar actualización
			updated, err := repo.GetOperationResult(helper.Context, operation.PK, operation.SK)
			assert.NoError(t, err)
			assert.Equal(t, "completed", updated.Status)
		})

		t.Run("should return error when updating non-existent operation", func(t *testing.T) {
			// act
			err := repo.UpdateOperationStatus(helper.Context, "non-existent", "test-song", "completed")

			// assert
			assert.Error(t, err)
			assert.ErrorIs(t, err, mongodb.ErrOperationNotFound)
		})

		t.Run("should return error with invalid parameters", func(t *testing.T) {
			// act
			err := repo.UpdateOperationStatus(helper.Context, "", "", "")

			// assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "operationID, songID y status son requeridos")
		})
	})

	t.Run("DeleteOperationResult", func(t *testing.T) {
		t.Run("should delete operation successfully", func(t *testing.T) {
			// arrange
			operation := &model.OperationResult{
				SK:     "test-song-5",
				Status: "pending",
			}

			err := repo.SaveOperationsResult(helper.Context, operation)
			assert.NoError(t, err)

			// act
			err = repo.DeleteOperationResult(helper.Context, operation.PK, operation.SK)

			// assert
			assert.NoError(t, err)

			// verificar que se elimino
			_, err = repo.GetOperationResult(helper.Context, operation.PK, operation.SK)
			assert.ErrorIs(t, err, mongodb.ErrOperationNotFound)

		})

		t.Run("should return error when deleting non-existent operation", func(t *testing.T) {
			// act
			err := repo.DeleteOperationResult(helper.Context, "non-existent", "test-song")

			// assert
			assert.Error(t, err)
			assert.ErrorIs(t, err, mongodb.ErrOperationNotFound)
		})

		t.Run("should return error with empty parameters", func(t *testing.T) {
			// act
			err := repo.DeleteOperationResult(helper.Context, "", "")

			// assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "id y songID son requeridos")
		})
	})

	t.Run("Integration flows", func(t *testing.T) {
		t.Run("should handle complete CRUD operation flow", func(t *testing.T) {
			// arrange
			operation := &model.OperationResult{
				SK:     "test-song-6",
				Status: "pending",
			}

			// create
			err := repo.SaveOperationsResult(helper.Context, operation)
			assert.NoError(t, err)
			assert.NotEmpty(t, operation.PK)

			// read
			retrieved, err := repo.GetOperationResult(helper.Context, operation.PK, operation.SK)
			assert.NoError(t, err)
			assert.Equal(t, operation.Status, retrieved.Status)

			// update
			err = repo.UpdateOperationStatus(helper.Context, operation.PK, operation.SK, "completed")
			assert.NoError(t, err)

			// verify update
			updated, err := repo.GetOperationResult(helper.Context, operation.PK, operation.SK)
			assert.NoError(t, err)
			assert.Equal(t, "completed", updated.Status)

			// delete
			err = repo.DeleteOperationResult(helper.Context, operation.PK, operation.SK)
			assert.NoError(t, err)

			// verificar delete
			_, err = repo.GetOperationResult(helper.Context, operation.PK, operation.SK)
			assert.ErrorIs(t, err, mongodb.ErrOperationNotFound)
		})

		t.Run("should handle concurrent operations", func(t *testing.T) {
			// Arrange
			operation1 := &model.OperationResult{
				SK:     "test-song-7",
				Status: "pending",
			}
			operation2 := &model.OperationResult{
				SK:     "test-song-8",
				Status: "pending",
			}

			// Act - Save both operations
			err1 := repo.SaveOperationsResult(helper.Context, operation1)
			err2 := repo.SaveOperationsResult(helper.Context, operation2)

			// Assert
			assert.NoError(t, err1)
			assert.NoError(t, err2)
			assert.NotEqual(t, operation1.PK, operation2.PK)

			// Cleanup
			_ = repo.DeleteOperationResult(helper.Context, operation1.PK, operation1.SK)
			_ = repo.DeleteOperationResult(helper.Context, operation2.PK, operation2.SK)
		})
	})
}
