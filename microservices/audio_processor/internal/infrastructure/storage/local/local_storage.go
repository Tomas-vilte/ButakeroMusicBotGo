package local

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	errorsApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalStorage struct {
	config *config.Config
	log    logger.Logger
}

func NewLocalStorage(cfg *config.Config, log logger.Logger) (*LocalStorage, error) {
	if err := os.MkdirAll(cfg.Storage.LocalConfig.BasePath, 0777); err != nil {
		return nil, errorsApp.ErrLocalDirectoryNotWritable.WithMessage(fmt.Sprintf("error creando directorio base %s: %v", cfg.Storage.LocalConfig.BasePath, err))
	}

	testFile := filepath.Join(cfg.Storage.LocalConfig.BasePath, ".write_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return nil, errorsApp.ErrLocalDirectoryNotWritable.WithMessage(fmt.Sprintf("el directorio %s no es escribible: %v", testFile, err))
	}

	defer func() {
		if err := os.Remove(testFile); err != nil {
			log.Error("Error al eliminar el archivo", zap.Error(err))
		}
	}()

	return &LocalStorage{
		config: cfg,
		log:    log,
	}, nil
}

func (l *LocalStorage) UploadFile(ctx context.Context, key string, body io.Reader) error {
	log := l.log.With(
		zap.String("component", "LocalStorage"),
		zap.String("method", "UploadFile"),
		zap.String("key", key),
	)

	select {
	case <-ctx.Done():
		log.Error("Contexto cancelado durante la subida del archivo", zap.Error(ctx.Err()))
		return errorsApp.ErrLocalUploadFailed.WithMessage(fmt.Sprintf("contexto cancelado durante la subida del archivo: %v", ctx.Err()))
	default:
	}

	if body == nil {
		log.Error("El cuerpo del archivo es nulo")
		return errorsApp.ErrLocalInvalidFile.WithMessage("el body no puede ser nulo")
	}

	if !strings.HasSuffix(key, ".dca") {
		key += ".dca"
	}

	fullPath := filepath.Join(l.config.Storage.LocalConfig.BasePath, "audio", key)
	log.Info("Subiendo archivo", zap.String("full_path", fullPath))

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		log.Error("Error creando directorio", zap.Error(err))
		return errorsApp.ErrLocalUploadFailed.WithMessage(fmt.Sprintf("error creando directorio para %s: %v", fullPath, err))
	}

	file, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Error("Error creando archivo", zap.Error(err))
		return errorsApp.ErrLocalUploadFailed.WithMessage(fmt.Sprintf("error creando archivo %s: %v", fullPath, err))
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Error("Error al cerrar el archivo", zap.Error(err))
		}
	}()

	buf := make([]byte, 32*1024)
	_, err = io.CopyBuffer(file, body, buf)
	if err != nil {
		log.Error("Error escribiendo archivo", zap.Error(err))
		return errorsApp.ErrLocalUploadFailed.WithMessage(fmt.Sprintf("error escribiendo archivo %s: %v", fullPath, err))
	}

	log.Info("Archivo subido exitosamente")
	return nil
}

func (l *LocalStorage) GetFileMetadata(ctx context.Context, key string) (*model.FileData, error) {
	log := l.log.With(
		zap.String("component", "LocalStorage"),
		zap.String("method", "GetFileMetadata"),
		zap.String("key", key),
	)

	select {
	case <-ctx.Done():
		log.Error("Contexto cancelado durante la obtención de metadatos", zap.Error(ctx.Err()))
		return nil, errorsApp.ErrLocalGetMetadataFailed.WithMessage(fmt.Sprintf("contexto cancelado durante la obtención de metadata: %v", ctx.Err()))
	default:
	}

	if !strings.HasSuffix(key, ".dca") {
		key += ".dca"
	}

	fullPath := filepath.Join(l.config.Storage.LocalConfig.BasePath, "audio", key)
	cleanPath := filepath.Clean(fullPath)

	absPath, err := filepath.Abs(cleanPath)
	if err != nil {
		log.Error("Error obteniendo ruta absoluta", zap.Error(err))
		return nil, errorsApp.ErrLocalGetMetadataFailed.WithMessage(fmt.Sprintf("error obteniendo ruta absoluta para %s: %v", cleanPath, err))
	}

	log.Info("Obteniendo metadatos del archivo", zap.String("full_path", absPath))

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Error("Archivo no encontrado", zap.Error(err))
			return nil, errorsApp.ErrLocalFileNotFound.WithMessage(fmt.Sprintf("archivo %s no encontrado: %v", key, err))
		}
		log.Error("Error obteniendo información del archivo", zap.Error(err))
		return nil, errorsApp.ErrLocalGetMetadataFailed.WithMessage(fmt.Sprintf("error obteniendo información del archivo %s: %v", key, err))
	}

	readableSize := FormatFileSize(fileInfo.Size())
	log.Info("Metadatos obtenidos exitosamente", zap.String("file_size", readableSize))

	return &model.FileData{
		FilePath: absPath,
		FileType: "audio/dca",
		FileSize: readableSize,
	}, nil
}

func (l *LocalStorage) GetFileContent(ctx context.Context, path string, key string) (io.ReadCloser, error) {
	log := l.log.With(
		zap.String("component", "LocalStorage"),
		zap.String("method", "GetFileContent"),
		zap.String("path", path),
		zap.String("key", key),
	)

	select {
	case <-ctx.Done():
		log.Error("Contexto cancelado durante la obtención del contenido del archivo", zap.Error(ctx.Err()))
		return nil, errorsApp.ErrLocalGetContentFailed.WithMessage(fmt.Sprintf("contexto cancelado durante la obtención del contenido del archivo: %v", ctx.Err()))
	default:
	}

	fullPath := filepath.Join(path, key)
	log.Info("Obteniendo contenido del archivo", zap.String("full_path", fullPath))

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Error("Archivo no encontrado", zap.Error(err))
			return nil, errorsApp.ErrLocalFileNotFound.WithMessage(fmt.Sprintf("archivo %s no encontrado: %v", fullPath, err))
		}
		log.Error("Error abriendo archivo", zap.Error(err))
		return nil, errorsApp.ErrLocalGetContentFailed.WithMessage(fmt.Sprintf("error abriendo archivo %s: %v", fullPath, err))
	}

	log.Info("Contenido del archivo obtenido exitosamente")
	return file, nil
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
