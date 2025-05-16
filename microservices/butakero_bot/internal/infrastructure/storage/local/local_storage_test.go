//go:build !integration

package local_storage

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/errors_app"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalStorage_GetAudio(t *testing.T) {
	tempDir := t.TempDir()
	mockLogger := new(logging.MockLogger)

	testContent := []byte("test audio content")
	testFile := filepath.Join(tempDir, "test.mp3")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Error creando archivo de prueba: %v", err)
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	t.Run("Caso exitoso - archivo válido", func(t *testing.T) {
		storage := NewLocalStorage(mockLogger)

		rc, err := storage.GetAudio(context.Background(), testFile)
		if err != nil {
			t.Fatalf("Error inesperado: %v", err)
		}
		defer func() {
			if err := rc.Close(); err != nil {
				t.Fatalf("Error cerrando el archivo: %v", err)
			}
		}()

		content, err := io.ReadAll(rc)
		if err != nil {
			t.Fatalf("Error leyendo contenido: %v", err)
		}

		if string(content) != string(testContent) {
			t.Errorf("Contenido incorrecto. Esperado: %s, Obtenido: %s", testContent, content)
		}
	})

	t.Run("Archivo no encontrado", func(t *testing.T) {
		storage := NewLocalStorage(mockLogger)

		_, err := storage.GetAudio(context.Background(), filepath.Join(tempDir, "inexistente.mp3"))
		assert.True(t, errors_app.IsAppErrorWithCode(err, errors_app.ErrCodeLocalFileNotFound))
	})

	t.Run("Contexto cancelado", func(t *testing.T) {
		storage := NewLocalStorage(mockLogger)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := storage.GetAudio(ctx, testFile)
		assert.True(t, errors_app.IsAppErrorWithCode(err, errors_app.ErrCodeLocalGetContentFailed))
	})
}
