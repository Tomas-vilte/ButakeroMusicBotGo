package uploader

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/ecs/process_audio/internal/config"
	"github.com/Tomas-vilte/GoMusicBot/ecs/process_audio/internal/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"
	"io"
)

type (
	Uploader interface {
		UploadToS3(ctx context.Context, audioData io.Reader, key string) error
		DownloadFromS3(ctx context.Context, key string) (*S3Object, error)
	}

	S3Object struct {
		Body io.ReadCloser
	}

	S3Uploader struct {
		s3Client S3ClientInterface
		s3Upload S3UploaderInterface
		logger   logging.Logger
		config   config.Config
	}

	// S3UploaderInterface define los métodos necesarios para cargar archivos a S3.
	S3UploaderInterface interface {
		UploadWithContext(aws.Context, *s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	}

	S3ClientInterface interface {
		GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error)
	}
)

// NewS3Uploader crea un nuevo S3Uploader usando la región especificada.
func NewS3Uploader(s3Client S3ClientInterface, s3Upload S3UploaderInterface, logger logging.Logger, cfg config.Config) *S3Uploader {
	return &S3Uploader{
		s3Client: s3Client,
		s3Upload: s3Upload,
		logger:   logger,
		config:   cfg,
	}
}

func (u *S3Uploader) UploadToS3(ctx context.Context, audioData io.Reader, key string) error {
	u.logger.Info("Iniciando carga de datos DCA a S3", zap.String("bucket", u.config.BucketName), zap.String("key", key))
	upParams := &s3manager.UploadInput{
		Bucket: aws.String(u.config.BucketName),
		Key:    aws.String("audio/" + key),
		Body:   audioData,
	}

	result, err := u.s3Upload.UploadWithContext(ctx, upParams)
	if err != nil {
		u.logger.Error("Error al subir los datos DCA a S3", zap.Error(err))
		return fmt.Errorf("error al subir los datos DCA a S3")
	}
	u.logger.Info("Datos DCA subidos exitosamente a S3", zap.Any("result", result))
	return nil
}

func (u *S3Uploader) DownloadFromS3(ctx context.Context, key string) (*S3Object, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(u.config.BucketName),
		Key:    aws.String(key),
	}

	result, err := u.s3Client.GetObjectWithContext(ctx, input)
	if err != nil {
		u.logger.Error("Error al descargar el archivo de S3", zap.Error(err), zap.String("key", key))
		return nil, fmt.Errorf("error al descargar el archivo de S3: %w", err)
	}

	return &S3Object{Body: result.Body}, nil
}
