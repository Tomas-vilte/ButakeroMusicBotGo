//go:build !integration

package local_storage

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/mock"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
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
		defer rc.Close()

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
		if !os.IsNotExist(err) {
			t.Errorf("Error esperado: %v, Obtenido: %v", os.ErrNotExist, err)
		}
	})

	t.Run("Contexto cancelado", func(t *testing.T) {
		storage := NewLocalStorage(mockLogger)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := storage.GetAudio(ctx, testFile)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Error esperado: %v, Obtenido: %v", context.Canceled, err)
		}
	})

	t.Run("Cierre del archivo por contexto", func(t *testing.T) {
		storage := NewLocalStorage(mockLogger)
		ctx, cancel := context.WithCancel(context.Background())

		rc, err := storage.GetAudio(ctx, testFile)
		if err != nil {
			t.Fatalf("Error inesperado: %v", err)
		}

		go func() {
			time.Sleep(100 * time.Millisecond)
			cancel()
		}()

		<-ctx.Done()
		time.Sleep(200 * time.Millisecond)

		_, err = rc.Read(make([]byte, 1))
		if err == nil {
			t.Error("Se esperaba error al leer archivo cerrado")
		}
	})
}
