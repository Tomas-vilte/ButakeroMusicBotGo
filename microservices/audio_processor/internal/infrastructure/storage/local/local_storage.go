package local

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	config *config.Config
}

func NewLocalStorage(cfg *config.Config) (*LocalStorage, error) {
	if err := os.MkdirAll(cfg.Storage.LocalConfig.BasePath, 0777); err != nil {
		return nil, fmt.Errorf("error creando directorio base %s:%w", cfg.Storage.LocalConfig.BasePath, err)
	}

	testFile := filepath.Join(cfg.Storage.LocalConfig.BasePath, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return nil, fmt.Errorf("el directorio %s no es escribible: %w", testFile, err)
	}

	defer func() {
		if err := os.Remove(testFile); err != nil {
			_ = fmt.Errorf("error al eliminar el archivo: %w", err)
		}
	}()

	return &LocalStorage{config: cfg}, nil
}

func (l *LocalStorage) UploadFile(ctx context.Context, key string, body io.Reader) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("contexto cancelado durante la subida del archivo: %w", ctx.Err())
	default:
	}

	if body == nil {
		return fmt.Errorf("el body no puede ser nulo")
	}

	if !strings.HasSuffix(key, ".dca") {
		key += ".dca"
	}

	fullPath := filepath.Join(l.config.Storage.LocalConfig.BasePath, "audio", key)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("error creando directorio para %s: %w", fullPath, err)
	}

	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("error creando archivo %s: %w", fullPath, err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			_ = fmt.Errorf("error al cerrar el archivo %s: %w", fullPath, err)
		}
	}()

	buf := make([]byte, 32*1024)
	_, err = io.CopyBuffer(file, body, buf)
	if err != nil {
		return fmt.Errorf("error escribiendo archivo %s: %w", fullPath, err)
	}
	return nil
}

func (l *LocalStorage) GetFileMetadata(ctx context.Context, key string) (*model.FileData, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("contexto cancelado durante la obtención de metadata: %w", ctx.Err())
	default:
	}

	if !strings.HasSuffix(key, ".dca") {
		key += ".dca"
	}

	fullPath := filepath.Join(l.config.Storage.LocalConfig.BasePath, "audio", key)

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("archivo %s no encontrado: %w", key, err)
		}
		return nil, fmt.Errorf("error obteniendo información del archivo %s: %w", key, err)
	}
	return &model.FileData{
		FilePath:  "audio/" + key,
		FileType:  "audio/dca",
		FileSize:  FormatFileSize(fileInfo.Size()),
		PublicURL: fmt.Sprintf("file://%s", fullPath),
	}, nil
}

func FormatFileSize(sizeBytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case sizeBytes >= GB:
		return fmt.Sprintf("%.2fGB", float64(sizeBytes)/float64(GB))
	case sizeBytes >= MB:
		return fmt.Sprintf("%.2fMB", float64(sizeBytes)/float64(MB))
	case sizeBytes >= KB:
		return fmt.Sprintf("%.2fKB", float64(sizeBytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", sizeBytes)
	}
}
