package s3_storage

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
	"io"
)

type S3Storage struct {
	client *s3.Client
	logger logging.Logger
	config *config.Config
}

// NewS3Storage crea una nueva instancia de S3Storage con validación de configuración.
func NewS3Storage(cfg *config.Config, logger logging.Logger) (*S3Storage, error) {
	if cfg == nil || cfg.Storage.S3Config.BucketName == "" {
		return nil, fmt.Errorf("configuración de S3 inválida")
	}

	awsConfig, err := awsCfg.LoadDefaultConfig(
		context.Background(),
		awsCfg.WithRegion(cfg.AWS.Region),
	)
	if err != nil {
		logger.Error("Error al cargar configuración de AWS",
			zap.Error(err),
			zap.String("region", cfg.AWS.Region))
		return nil, fmt.Errorf("error de configuración AWS: %w", err)
	}

	return &S3Storage{
		client: s3.NewFromConfig(awsConfig),
		config: cfg,
		logger: logger,
	}, nil
}

// GetAudio obtiene un archivo de audio desde S3 con validación de parámetros y manejo contextual.
func (s *S3Storage) GetAudio(ctx context.Context, songPath string) (io.ReadCloser, error) {
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

	bucket := s.config.Storage.S3Config.BucketName
	logger.Debug("Obteniendo audio de S3",
		zap.String("bucket", bucket),
		zap.String("path", songPath))

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(songPath),
	})

	if err != nil {
		logger.Error("Error al obtener objeto de S3",
			zap.String("bucket", bucket),
			zap.String("path", songPath),
			zap.Error(err))
		return nil, fmt.Errorf("error al obtener de S3: %w", err)
	}

	logger.Debug("Audio obtenido exitosamente",
		zap.String("bucket", bucket),
		zap.String("path", songPath))

	return result.Body, nil
}
