package s3uploader

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"io"
	"os"
)

// Uploader interface define los métodos necesarios para cargar archivos o contenido a S3.
type Uploader interface {
	UploadFile(ctx context.Context, filePath, bucketName, key string) error
	UploadContent(ctx context.Context, content io.Reader, bucketName, key string) error
}

// S3Uploader implementa la interfaz Uploader usando el cliente S3.
type S3Uploader struct {
	S3Client s3iface.S3API
}

// NewS3Uploader crea un nuevo S3Uploader usando la región especificada.
func NewS3Uploader(region, accessKey, secretAccessKey string) (*S3Uploader, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewStaticCredentials(accessKey, secretAccessKey, ""),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error al crear la sesión AWS: %v", err)
	}
	svc := s3.New(sess)
	return &S3Uploader{S3Client: svc}, nil
}

// NewS3UploaderWithClient crea un nuevo S3Uploader usando un cliente S3 personalizado.
func NewS3UploaderWithClient(client s3iface.S3API) *S3Uploader {
	return &S3Uploader{S3Client: client}
}

// UploadFile carga un archivo desde el sistema de archivos local a S3.
func (u *S3Uploader) UploadFile(ctx context.Context, filePath, bucketName, key string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	return u.UploadContent(ctx, file, bucketName, key)
}

// UploadContent carga contenido proporcionado como io.Reader a S3.
func (u *S3Uploader) UploadContent(ctx context.Context, content io.Reader, bucketName, key string) error {
	_, err := u.S3Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
		Body:   aws.ReadSeekCloser(content),
	})
	if err != nil {
		return fmt.Errorf("error al subir el contenido a S3: %v", err)
	}
	return nil
}
