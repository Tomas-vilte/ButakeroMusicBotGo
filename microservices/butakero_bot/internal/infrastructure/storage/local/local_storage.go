package local_storage

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	config  *config.Config
	logger  logging.Logger
	baseDir string
}

// NewLocalStorage crea una instancia de LocalStorage con validación de directorio.
func NewLocalStorage(cfg *config.Config, logger logging.Logger) (*LocalStorage, error) {
	if cfg == nil || cfg.Storage.LocalConfig.Directory == "" {
		return nil, fmt.Errorf("configuración local inválida")
	}

	absPath, err := filepath.Abs(cfg.Storage.LocalConfig.Directory)
	if err != nil {
		logger.Error("Error al obtener ruta absoluta",
			zap.String("directorio", cfg.Storage.LocalConfig.Directory),
			zap.Error(err))
		return nil, fmt.Errorf("error de directorio: %w", err)
	}

	return &LocalStorage{
		config:  cfg,
		logger:  logger,
		baseDir: absPath,
	}, nil
}

// GetAudio obtiene un archivo de audio local con prevención de directory traversal.
func (s *LocalStorage) GetAudio(ctx context.Context, songPath string) (io.ReadCloser, error) {
	if songPath == "" {
		return nil, fmt.Errorf("songPath no puede estar vacío")
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	fullPath := filepath.Join(s.baseDir, filepath.Clean("/"+songPath))
	relPath, err := filepath.Rel(s.baseDir, fullPath)
	if err != nil || strings.HasPrefix(relPath, "..") {
		s.logger.Error("Intento de acceso fuera del directorio base",
			zap.String("path", fullPath))
		return nil, fmt.Errorf("ruta no permitida")
	}

	s.logger.Debug("Obteniendo audio local",
		zap.String("path", fullPath))

	file, err := os.Open(fullPath)
	if err != nil {
		s.logger.Error("Error al abrir archivo local",
			zap.String("path", fullPath),
			zap.Error(err))
		return nil, fmt.Errorf("error al abrir archivo: %w", err)
	}

	return file, nil
}
