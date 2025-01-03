package cloud

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
)

type (
	// S3Client define la interfaz para interactuar con el servicio S3 de AWS.
	// Permite subir archivos y obtener información del encabezado del objeto.
	S3Client interface {
		// PutObject sube un objeto a un bucket de S3.
		PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)

		// HeadObject obtiene la información del encabezado del objeto de S3.
		HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)

		// GetObject obtiene el contenido del objeto de S3.
		GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	}
)

// S3Storage implementa la interfaz Storage utilizando el servicio S3 de AWS.
// Permite subir archivos y obtener metadatos de archivos almacenados en S3.
type S3Storage struct {
	// Client es el cliente de S3 utilizado para interactuar con el servicio.
	Client S3Client
	// Config es la configuración de la aplicación.
	Config *config.Config
}

// NewS3Storage crea una nueva instancia de S3Storage.
// Configura el cliente de S3 con las credenciales y la región especificadas en la configuración.
func NewS3Storage(cfgApplication *config.Config) (*S3Storage, error) {
	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(), awsCfg.WithRegion(cfgApplication.AWS.Region))
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Storage{
		Client: client,
		Config: cfgApplication,
	}, nil
}

// UploadFile sube un archivo al bucket de S3 con la clave especificada.
// El archivo se sube con la ruta "audio/" concatenada con la clave.
func (s *S3Storage) UploadFile(ctx context.Context, key string, body io.Reader) error {
	if body == nil {
		return fmt.Errorf("el cuerpo no puede ser nulo")
	}
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.Config.Storage.S3Config.BucketName),
		Key:    aws.String("audio/" + key),
		Body:   body,
	}

	_, err := s.Client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("error subiendo archivo a S3: %w", err)
	}
	return nil
}

// GetFileMetadata obtiene los metadatos del archivo subido a S3 y devuelve un model.FileData.
func (s *S3Storage) GetFileMetadata(ctx context.Context, key string) (*model.FileData, error) {
	headInput := &s3.HeadObjectInput{
		Bucket: aws.String(s.Config.Storage.S3Config.BucketName),
		Key:    aws.String("audio/" + key),
	}

	headResult, err := s.Client.HeadObject(ctx, headInput)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo metadata del archivo de S3: %w", err)
	}

	readableSize := formatFileSize(*headResult.ContentLength)

	return &model.FileData{
		FilePath:  "audio/" + key,
		FileType:  *headResult.ContentType,
		FileSize:  readableSize,
		PublicURL: fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.Config.Storage.S3Config.BucketName, key),
	}, nil
}

// formatFileSize formatea el tamaño del archivo en una representación legible.
func formatFileSize(sizeBytes int64) string {
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

// GetFileContent obtiene el contenido del archivo con la clave especificada.
func (s *S3Storage) GetFileContent(ctx context.Context, path string, key string) (io.ReadCloser, error) {
	getInput := &s3.GetObjectInput{
		Bucket: aws.String(s.Config.Storage.S3Config.BucketName),
		Key:    aws.String(path + key),
	}

	getResult, err := s.Client.GetObject(ctx, getInput)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo contenido del archivo %s de S3: %w", key, err)
	}

	return getResult.Body, nil
}
