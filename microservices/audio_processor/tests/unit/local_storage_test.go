package unit

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/storage/local"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLocalStorage(t *testing.T) {
	setupTest := func(t *testing.T) (*local.LocalStorage, string) {
		tempDir, err := os.MkdirTemp("", "storage-test-*")
		assert.NoError(t, err, "Error creando directorio temporal")

		storage, err := local.NewLocalStorage(tempDir)
		assert.NoError(t, err, "Error creando LocalStorage")

		t.Cleanup(func() {
			if err := os.RemoveAll(tempDir); err != nil {
				t.Fatalf("Error al eliminar el directorio temporal: %s", err)
			}
		})
		return storage, tempDir
	}

	t.Run("NewLocalStorage", func(t *testing.T) {
		t.Run("should create base directory if it doesn't exist", func(t *testing.T) {
			// arrange
			tempDir := filepath.Join(os.TempDir(), "storage-test-new")
			defer func() {
				if err := os.RemoveAll(tempDir); err != nil {
					t.Fatalf("Error al eliminar el directorio temporal: %s", err)
				}
			}()

			// act
			storage, err := local.NewLocalStorage(tempDir)

			// assert
			assert.NoError(t, err)
			assert.NotNil(t, storage)
			assert.DirExists(t, tempDir)
		})

		t.Run("should fail if you don't have write permissions", func(t *testing.T) {
			// arrange
			tempDir := filepath.Join(os.TempDir(), "storage-test-readonly")

			if err := os.MkdirAll(tempDir, 0555); err != nil { // solo lectura
				t.Fatalf("Error al creando el directorio temporal: %s", err)
			}

			defer func() {
				if err := os.RemoveAll(tempDir); err != nil {
					t.Fatalf("Error al eliminar el directorio temporal: %s", err)
				}
			}()

			// act
			storage, err := local.NewLocalStorage(tempDir)

			// assert
			assert.Error(t, err)
			assert.Nil(t, storage)
			assert.Contains(t, err.Error(), "no es escribible")
		})
	})

	t.Run("UploadFile", func(t *testing.T) {
		t.Run("should save DCA file correctly", func(t *testing.T) {
			// arrange
			storage, tempDir := setupTest(t)
			content := "contenido de prueba"
			key := ".dca"
			ctx := context.Background()

			// act
			err := storage.UploadFile(ctx, key, strings.NewReader(content))

			// assert
			assert.NoError(t, err)
			expectedPath := filepath.Join(tempDir, "audio", key)
			assert.FileExists(t, expectedPath)

			// verificar contenido
			savedContent, err := os.ReadFile(expectedPath)
			assert.NoError(t, err)
			assert.Equal(t, content, string(savedContent))
		})

		t.Run("should add .dca extension if missing", func(t *testing.T) {
			// arrange
			storage, tempDir := setupTest(t)
			content := "contenido de prueba"
			key := "test"
			ctx := context.Background()

			// act
			err := storage.UploadFile(ctx, key, strings.NewReader(content))

			// assert
			assert.NoError(t, err)
			expectedPath := filepath.Join(tempDir, "audio", key+".dca")
			assert.FileExists(t, expectedPath)
		})

		t.Run("should handle error when context is canceled", func(t *testing.T) {
			// arrange
			storage, _ := setupTest(t)
			content := "contenido de prueba"
			key := ".dca"
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			// act
			err := storage.UploadFile(ctx, key, strings.NewReader(content))

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "contexto cancelado durante la subida del archivo")
		})

		t.Run("should handle body null", func(t *testing.T) {
			// arrange
			storage, _ := setupTest(t)
			ctx := context.Background()

			// act
			err := storage.UploadFile(ctx, "test.dca", nil)

			// assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "el body no puede ser nulo")
		})

		t.Run("should handle large files correctly", func(t *testing.T) {
			// arrange
			storage, _ := setupTest(t)
			largeContent := bytes.Repeat([]byte("a"), 1024*1024) // 1MB de datos
			ctx := context.Background()

			// act
			err := storage.UploadFile(ctx, "large.dca", bytes.NewReader(largeContent))

			// assert
			assert.NoError(t, err)
		})
	})

	t.Run("GetFileMetadata", func(t *testing.T) {
		t.Run("should get metadata correctly", func(t *testing.T) {
			// arrange
			storage, _ := setupTest(t)
			content := "contenido de prueba"
			key := "test.dca"
			ctx := context.Background()

			err := storage.UploadFile(ctx, key, strings.NewReader(content))
			assert.NoError(t, err)

			// act
			metadata, err := storage.GetFileMetadata(ctx, key)

			// assert
			assert.NoError(t, err)
			assert.NotNil(t, metadata)
			assert.Equal(t, "audio/"+key, metadata.FilePath)
			assert.Equal(t, "audio/dca", metadata.FileType)
			assert.Contains(t, metadata.FileSize, "B") // Debería tener el formato correcto
		})

		t.Run("should handle non-existing file", func(t *testing.T) {
			// arrange
			storage, _ := setupTest(t)
			ctx := context.Background()

			// act
			metadata, err := storage.GetFileMetadata(ctx, "no-exist.dca")

			// assert
			assert.Error(t, err)
			assert.Nil(t, metadata)
			assert.Contains(t, err.Error(), "no encontrado")

		})

		t.Run("should add .dca extension if missing", func(t *testing.T) {
			// Arrange
			storage, _ := setupTest(t)
			content := "contenido de prueba"
			ctx := context.Background()

			err := storage.UploadFile(ctx, "test", strings.NewReader(content))
			assert.NoError(t, err)

			// Act
			metadata, err := storage.GetFileMetadata(ctx, "test")

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, metadata)
			assert.Equal(t, "audio/test.dca", metadata.FilePath)
		})

		t.Run("should handle canceled context", func(t *testing.T) {
			// Arrange
			storage, _ := setupTest(t)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			// Act
			metadata, err := storage.GetFileMetadata(ctx, "test.dca")

			// Assert
			assert.Error(t, err)
			assert.Nil(t, metadata)
			assert.Contains(t, err.Error(), "contexto cancelado")
		})
	})

	t.Run("formatFileSize", func(t *testing.T) {
		tests := []struct {
			name     string
			size     int64
			expected string
		}{
			{"bytes", 500, "500B"},
			{"kilobytes", 1024 * 2, "2.00KB"},
			{"megabytes", 1024 * 1024 * 3, "3.00MB"},
			{"gigabytes", 1024 * 1024 * 1024 * 4, "4.00GB"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Act
				result := local.FormatFileSize(tt.size)
				// Assert
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	// arrange
	storage, _ := setupTest(t)
	ctx := context.Background()
	numGoroutines := 10
	done := make(chan bool)

	// act
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			key := fmt.Sprintf("concurrent_%d.dca", index)
			content := fmt.Sprintf("content_%d", index)

			err := storage.UploadFile(ctx, key, strings.NewReader(content))
			assert.NoError(t, err)

			metadata, err := storage.GetFileMetadata(ctx, key)
			assert.NoError(t, err)
			assert.NotNil(t, metadata)

			done <- true
		}(i)
	}

	// assert
	for i := 0; i < numGoroutines; i++ {
		select {
		case <-done:
			continue
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout esperando por goroutines")
		}
	}
}

func setupTest(t *testing.T) (*local.LocalStorage, string) {
	tempDir, err := os.MkdirTemp("", "storage-test-*")
	require.NoError(t, err, "Error creando directorio temporal")

	storage, err := local.NewLocalStorage(tempDir)
	require.NoError(t, err, "Error creando LocalStorage")

	t.Cleanup(func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Fatalf("Error al eliminar directorio: %s", err)
		}
	})

	return storage, tempDir
}
