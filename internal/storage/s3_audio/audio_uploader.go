package s3_audio

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/config"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"
	"io"
)

type (

	// Uploader interface define los métodos necesarios para cargar archivos o contenido a S3.
	Uploader interface {
		UploadDCA(ctx context.Context, audioData io.Reader, key string) error
		FileExists(ctx context.Context, key string) (bool, error)
		DownloadDCA(ctx context.Context, key string) (io.Reader, error)
	}

	// S3UploaderInterface define los métodos necesarios para cargar archivos a S3.
	S3UploaderInterface interface {
		UploadWithContext(aws.Context, *s3manager.UploadInput, ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
	}

	// S3DownloaderInterface define los métodos necesarios para descargar archivos de S3.
	S3DownloaderInterface interface {
		DownloadWithContext(aws.Context, io.WriterAt, *s3.GetObjectInput, ...func(*s3manager.Downloader)) (int64, error)
	}

	// S3ClientInterface define los métodos necesarios para interactuar con el cliente S3.
	S3ClientInterface interface {
		HeadObjectWithContext(ctx aws.Context, input *s3.HeadObjectInput, opts ...request.Option) (*s3.HeadObjectOutput, error)
	}
)

// S3Uploader implementa la interfaz Uploader usando el cliente S3.
type S3Uploader struct {
	S3Uploader   S3UploaderInterface
	S3Downloader S3DownloaderInterface
	S3Client     S3ClientInterface
	Logger       logging.Logger
	Config       *config.Config
}

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
	downloader := s3manager.NewDownloader(sess)
	client := s3.New(sess)
	return &S3Uploader{
		S3Uploader:   uploader,
		S3Downloader: downloader,
		S3Client:     client,
		Logger:       logger,
		Config:       &credentialsAws,
	}, nil
}

// UploadDCA carga los datos DCA desde audioData a S3.
func (u *S3Uploader) UploadDCA(ctx context.Context, audioData io.Reader, key string) error {
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

func (u *S3Uploader) FileExists(ctx context.Context, key string) (bool, error) {
	_, err := u.S3Client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(u.Config.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (u *S3Uploader) DownloadDCA(ctx context.Context, key string) (io.Reader, error) {
	buff := &aws.WriteAtBuffer{}
	_, err := u.S3Downloader.DownloadWithContext(ctx, buff, &s3.GetObjectInput{
		Bucket: aws.String(u.Config.BucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(buff.Bytes()), nil
}

func isNotFoundError(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == "NotFound" || awsErr.Code() == s3.ErrCodeNoSuchKey {
			return true
		}
	}
	return false
}
