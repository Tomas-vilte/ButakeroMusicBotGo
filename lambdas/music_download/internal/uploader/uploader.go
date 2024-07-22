package uploader

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/config"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"
	"io"
)

type (
	Uploader interface {
		UploadToS3(ctx context.Context, audioData io.Reader, key string) error
	}

	S3Uploader struct {
		svc        *s3.S3
		S3Uploader S3UploaderInterface
		Logger     logging.Logger
		Config     config.Config
	}

	// S3UploaderInterface define los métodos necesarios para cargar archivos a S3.
	S3UploaderInterface interface {
		UploadWithContext(aws.Context, *s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	}
)

// NewS3Uploader crea un nuevo S3Uploader usando la región especificada.
func NewS3Uploader(logger logging.Logger, credentialsAws config.Config) (*S3Uploader, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(credentialsAws.Region),
			Credentials: credentials.NewStaticCredentials(credentialsAws.AccessKey, credentialsAws.SecretKey, ""),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error al crear la session de AWS: %w", err)
	}
	uploader := s3manager.NewUploader(sess)
	client := s3.New(sess)
	return &S3Uploader{
		S3Uploader: uploader,
		svc:        client,
		Logger:     logger,
		Config:     credentialsAws,
	}, nil
}

func (u *S3Uploader) UploadToS3(ctx context.Context, audioData io.Reader, key string) error {
	u.Logger.Info("Iniciando carga de datos DCA a S3", zap.String("bucket", u.Config.BucketName), zap.String("key", key))

	upParams := &s3manager.UploadInput{
		Bucket: aws.String(u.Config.BucketName),
		Key:    aws.String(key),
		Body:   audioData,
	}

	result, err := u.S3Uploader.UploadWithContext(ctx, upParams)
	if err != nil {
		u.Logger.Error("Error al subir los datos DCA a S3", zap.Error(err))
		return fmt.Errorf("error al subir los datos DCA a S3")
	}
	u.Logger.Info("Datos DCA subidos exitosamente a S3", zap.Any("result", result))
	return nil
}
