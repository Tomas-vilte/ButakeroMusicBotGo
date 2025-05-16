package local_storage

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/errors_app"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
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
		return nil, errors_app.NewAppError(
			errors_app.ErrCodeInvalidInput,
			"songPath no puede estar vacío",
			nil,
		)
	}

	logger := s.logger.With(
		zap.String("component", "LocalStorage"),
		zap.String("trace_id", trace.GetTraceID(ctx)),
		zap.String("method", "GetAudio"),
		zap.String("songPath", songPath),
	)

	select {
	case <-ctx.Done():
		logger.Debug("Contexto cancelado antes de abrir el archivo")
		return nil, errors_app.NewAppError(
			errors_app.ErrCodeLocalGetContentFailed,
			"contexto cancelado durante la obtención del archivo",
			ctx.Err(),
		)
	default:
	}

	fullPath := filepath.Clean(songPath)

	if strings.Contains(fullPath, "..") {
		logger.Error("Intento de acceso fuera del directorio permitido",
			zap.String("path", fullPath))
		return nil, errors_app.NewAppError(
			errors_app.ErrCodeLocalInvalidFile,
			"ruta no permitida",
			nil,
		)
	}

	logger.Debug("Obteniendo audio local",
		zap.String("path", fullPath))

	file, err := os.Open(fullPath)
	if err != nil {
		logger.Error("Error al abrir archivo local",
			zap.String("path", fullPath),
			zap.Error(err))
		if os.IsNotExist(err) {
			return nil, errors_app.NewAppError(
				errors_app.ErrCodeLocalFileNotFound,
				"archivo no encontrado",
				err,
			)
		}

		return nil, errors_app.NewAppError(
			errors_app.ErrCodeLocalGetContentFailed,
			"error al abrir archivo",
			err,
		)
	}
	return file, nil
}
