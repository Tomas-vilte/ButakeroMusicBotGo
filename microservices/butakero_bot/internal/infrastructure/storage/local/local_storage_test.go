package local_storage

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLocalStorage_GetAudio(t *testing.T) {
	// Configuración base para las pruebas
	tempDir := t.TempDir()
	validCfg := &config.Config{
		Storage: config.StorageConfig{
			LocalConfig: config.LocalConfig{
				Directory: tempDir,
			},
		},
	}
	logger, err := logging.NewZapLogger()
	require.NoError(t, err)

	testContent := []byte("test audio content")
	testFile := filepath.Join(tempDir, "test.mp3")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Error creando archivo de prueba: %v", err)
	}

	t.Run("Caso exitoso - archivo válido", func(t *testing.T) {
		storage, err := NewLocalStorage(validCfg, logger)
		if err != nil {
			t.Fatalf("Error creando storage: %v", err)
		}

		rc, err := storage.GetAudio(context.Background(), "test.mp3")
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
		storage, _ := NewLocalStorage(validCfg, logger)

		_, err := storage.GetAudio(context.Background(), "inexistente.mp3")
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("Error esperado: %v, Obtenido: %v", os.ErrNotExist, err)
		}
	})
	
	t.Run("Contexto cancelado", func(t *testing.T) {
		storage, _ := NewLocalStorage(validCfg, logger)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := storage.GetAudio(ctx, "test.mp3")
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Error esperado: %v, Obtenido: %v", context.Canceled, err)
		}
	})

	t.Run("Cierre del archivo por contexto", func(t *testing.T) {
		storage, _ := NewLocalStorage(validCfg, logger)
		ctx, cancel := context.WithCancel(context.Background())

		rc, err := storage.GetAudio(ctx, "test.mp3")
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
