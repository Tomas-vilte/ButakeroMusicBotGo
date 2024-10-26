package integration

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/repository/mongodb"
	mongoHelper "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/tests/testutil/mongodb"
	"github.com/google/uuid"
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

	// Función helper para generar UUIDs válidos
	generateValidUUID := func() string {
		return uuid.New().String()
	}

	t.Run("NewOperationRepository", func(t *testing.T) {
		t.Run("should create repository successfully", func(t *testing.T) {
			repo, err := mongodb.NewOperationRepository(mongodb.OperationOptions{
				Collection: helper.MongoDB.GetCollection("operations"),
				Log:        helper.Logger,
			})

			assert.NoError(t, err)
			assert.NotNil(t, repo)
		})

		t.Run("should return error when collection is nil", func(t *testing.T) {
			repo, err := mongodb.NewOperationRepository(mongodb.OperationOptions{
				Collection: nil,
				Log:        helper.Logger,
			})

			assert.Error(t, err)
			assert.Nil(t, repo)
			assert.Contains(t, err.Error(), "collection no puede ser nil")
		})

		t.Run("should return error when logger is nil", func(t *testing.T) {
			repo, err := mongodb.NewOperationRepository(mongodb.OperationOptions{
				Collection: helper.MongoDB.GetCollection("operations"),
				Log:        nil,
			})

			assert.Error(t, err)
			assert.Nil(t, repo)
			assert.Contains(t, err.Error(), "logger no puede ser nil")
		})
	})

	t.Run("SaveOperationResult", func(t *testing.T) {
		t.Run("should save operation successfully with valid UUIDs", func(t *testing.T) {
			operation := &model.OperationResult{
				PK:     generateValidUUID(),
				SK:     generateValidUUID(),
				Status: "pending",
			}

			err := repo.SaveOperationsResult(helper.Context, operation)

			assert.NoError(t, err)
			assert.NotEmpty(t, operation.PK)

			saved, err := repo.GetOperationResult(helper.Context, operation.PK, operation.SK)
			assert.NoError(t, err)
			assert.Equal(t, operation.Status, saved.Status)
			assert.Equal(t, operation.SK, saved.SK)
		})

		t.Run("should return error when operation has invalid UUID", func(t *testing.T) {
			operation := &model.OperationResult{
				PK:     "invalid-uuid",
				SK:     generateValidUUID(),
				Status: "pending",
			}

			err := repo.SaveOperationsResult(helper.Context, operation)
			assert.ErrorIs(t, err, mongodb.ErrInvalidUUID)
		})

		t.Run("should return error when operation is nil", func(t *testing.T) {
			err := repo.SaveOperationsResult(helper.Context, nil)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "result no puede ser nil")
		})

		t.Run("should generate valid PK when not provided", func(t *testing.T) {
			operation := &model.OperationResult{
				SK:     generateValidUUID(),
				Status: "pending",
			}

			err := repo.SaveOperationsResult(helper.Context, operation)
			assert.NoError(t, err)
			assert.NotEmpty(t, operation.PK)

			// Verificar que el PK generado es un UUID válido
			_, err = uuid.Parse(operation.PK)
			assert.NoError(t, err)
		})
	})

	t.Run("GetOperationResult", func(t *testing.T) {
		t.Run("should get existing operation successfully with valid UUIDs", func(t *testing.T) {
			operation := &model.OperationResult{
				PK:     generateValidUUID(),
				SK:     generateValidUUID(),
				Status: "pending",
			}

			err := repo.SaveOperationsResult(helper.Context, operation)
			assert.NoError(t, err)

			result, err := repo.GetOperationResult(helper.Context, operation.PK, operation.SK)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, operation.Status, result.Status)
			assert.Equal(t, operation.SK, result.SK)
		})

		t.Run("should return error with invalid UUID", func(t *testing.T) {
			result, err := repo.GetOperationResult(helper.Context, "invalid-uuid", generateValidUUID())
			assert.ErrorIs(t, err, mongodb.ErrInvalidUUID)
			assert.Nil(t, result)
		})

		t.Run("should return ErrOperationNotFound for non-existent operation", func(t *testing.T) {
			result, err := repo.GetOperationResult(helper.Context, generateValidUUID(), generateValidUUID())
			assert.ErrorIs(t, err, mongodb.ErrOperationNotFound)
			assert.Nil(t, result)
		})
	})

	t.Run("UpdateOperationStatus", func(t *testing.T) {
		t.Run("should update status successfully with valid status", func(t *testing.T) {
			operation := &model.OperationResult{
				PK:     generateValidUUID(),
				SK:     generateValidUUID(),
				Status: "pending",
			}

			err := repo.SaveOperationsResult(helper.Context, operation)
			assert.NoError(t, err)

			err = repo.UpdateOperationStatus(helper.Context, operation.PK, operation.SK, "complete")
			assert.NoError(t, err)

			updated, err := repo.GetOperationResult(helper.Context, operation.PK, operation.SK)
			assert.NoError(t, err)
			assert.Equal(t, "complete", updated.Status)
		})

		t.Run("should return error with invalid status", func(t *testing.T) {
			operation := &model.OperationResult{
				PK:     generateValidUUID(),
				SK:     generateValidUUID(),
				Status: "pending",
			}

			err := repo.SaveOperationsResult(helper.Context, operation)
			assert.NoError(t, err)

			err = repo.UpdateOperationStatus(helper.Context, operation.PK, operation.SK, "invalid-status")
			assert.ErrorIs(t, err, mongodb.ErrInvalidStatus)
		})

		t.Run("should return error with invalid UUID", func(t *testing.T) {
			err := repo.UpdateOperationStatus(helper.Context, "invalid-uuid", generateValidUUID(), "complete")
			assert.ErrorIs(t, err, mongodb.ErrInvalidUUID)
		})
	})

	t.Run("DeleteOperationResult", func(t *testing.T) {
		t.Run("should delete operation successfully with valid UUIDs", func(t *testing.T) {
			operation := &model.OperationResult{
				PK:     generateValidUUID(),
				SK:     generateValidUUID(),
				Status: "pending",
			}

			err := repo.SaveOperationsResult(helper.Context, operation)
			assert.NoError(t, err)

			err = repo.DeleteOperationResult(helper.Context, operation.PK, operation.SK)
			assert.NoError(t, err)

			_, err = repo.GetOperationResult(helper.Context, operation.PK, operation.SK)
			assert.ErrorIs(t, err, mongodb.ErrOperationNotFound)
		})

		t.Run("should return error with invalid UUID", func(t *testing.T) {
			err := repo.DeleteOperationResult(helper.Context, "invalid-uuid", generateValidUUID())
			assert.ErrorIs(t, err, mongodb.ErrInvalidUUID)
		})
	})

	t.Run("Integration flows", func(t *testing.T) {
		t.Run("should handle complete CRUD operation flow with valid data", func(t *testing.T) {
			operation := &model.OperationResult{
				PK:     generateValidUUID(),
				SK:     generateValidUUID(),
				Status: "pending",
			}

			// Create
			err := repo.SaveOperationsResult(helper.Context, operation)
			assert.NoError(t, err)

			// Read
			retrieved, err := repo.GetOperationResult(helper.Context, operation.PK, operation.SK)
			assert.NoError(t, err)
			assert.Equal(t, operation.Status, retrieved.Status)

			// Update with valid status
			err = repo.UpdateOperationStatus(helper.Context, operation.PK, operation.SK, "complete")
			assert.NoError(t, err)

			// Verify update
			updated, err := repo.GetOperationResult(helper.Context, operation.PK, operation.SK)
			assert.NoError(t, err)
			assert.Equal(t, "complete", updated.Status)

			// Delete
			err = repo.DeleteOperationResult(helper.Context, operation.PK, operation.SK)
			assert.NoError(t, err)

			// Verify deletion
			_, err = repo.GetOperationResult(helper.Context, operation.PK, operation.SK)
			assert.ErrorIs(t, err, mongodb.ErrOperationNotFound)
		})

		t.Run("should handle concurrent operations with valid data", func(t *testing.T) {
			operation1 := &model.OperationResult{
				PK:     generateValidUUID(),
				SK:     generateValidUUID(),
				Status: "pending",
			}
			operation2 := &model.OperationResult{
				PK:     generateValidUUID(),
				SK:     generateValidUUID(),
				Status: "pending",
			}

			err1 := repo.SaveOperationsResult(helper.Context, operation1)
			err2 := repo.SaveOperationsResult(helper.Context, operation2)

			assert.NoError(t, err1)
			assert.NoError(t, err2)
			assert.NotEqual(t, operation1.PK, operation2.PK)

			// Cleanup
			_ = repo.DeleteOperationResult(helper.Context, operation1.PK, operation1.SK)
			_ = repo.DeleteOperationResult(helper.Context, operation2.PK, operation2.SK)
		})
	})
}
