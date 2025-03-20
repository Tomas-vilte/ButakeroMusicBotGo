package local_storage

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	logger logging.Logger
}

// NewLocalStorage crea una instancia de LocalStorage.
func NewLocalStorage(logger logging.Logger) *LocalStorage {
	return &LocalStorage{
		logger: logger,
	}
}

// GetAudio obtiene un archivo de audio local con prevención de directory traversal.
func (s *LocalStorage) GetAudio(ctx context.Context, songPath string) (io.ReadCloser, error) {
	if songPath == "" {
		return nil, fmt.Errorf("songPath no puede estar vacío")
	}

	logger := s.logger.With(
		zap.String("method", "GetAudio"),
		zap.String("songPath", songPath),
	)

	select {
	case <-ctx.Done():
		logger.Debug("Contexto cancelado antes de abrir el archivo")
		return nil, ctx.Err()
	default:
	}

	fullPath := filepath.Clean(songPath)

	if strings.Contains(fullPath, "..") {
		logger.Error("Intento de acceso fuera del directorio permitido",
			zap.String("path", fullPath))
		return nil, fmt.Errorf("ruta no permitida")
	}

	logger.Debug("Obteniendo audio local",
		zap.String("path", fullPath))

	file, err := os.Open(fullPath)
	if err != nil {
		logger.Error("Error al abrir archivo local",
			zap.String("path", fullPath),
			zap.Error(err))
		return nil, fmt.Errorf("error al abrir archivo: %w", err)
	}

	go func() {
		<-ctx.Done()
		logger.Debug("Contexto cancelado, cerrando archivo")
		if err := file.Close(); err != nil {
			logger.Error("Hubo un error al cerrar el archivo: ", zap.Error(err))
		}
	}()

	return file, nil
}
