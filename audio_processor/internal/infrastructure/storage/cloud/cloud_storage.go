package cloud

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	errorsApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
	"io"
)

type (
	// S3Client define la interfaz para interactuar con el servicio S3 de AWS.
	S3Client interface {
		PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
		HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error)
		GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	}
)

type S3Storage struct {
	Client S3Client
	Config *config.Config
	log    logger.Logger
}

func NewS3Storage(cfgApplication *config.Config, logger logger.Logger) (*S3Storage, error) {
	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(), awsCfg.WithRegion(cfgApplication.AWS.Region))
	if err != nil {
		return nil, errorsApp.ErrCodeDBConnectionFailed.WithMessage(fmt.Sprintf("error cargando configuración AWS: %v", err))
	}

	client := s3.NewFromConfig(cfg)

	return &S3Storage{
		Client: client,
		Config: cfgApplication,
		log:    logger,
	}, nil
}

func (s *S3Storage) UploadFile(ctx context.Context, key string, body io.Reader) error {
	log := s.log.With(
		zap.String("component", "S3Storage"),
		zap.String("method", "UploadFile"),
		zap.String("key", key),
	)

	if body == nil {
		log.Error("El cuerpo del archivo es nulo")
		return errorsApp.ErrS3InvalidFile.WithMessage("el cuerpo no puede ser nulo")
	}

	input := &s3.PutObjectInput{
		Bucket: aws.String(s.Config.Storage.S3Config.BucketName),
		Key:    aws.String("audio/" + key),
		Body:   body,
	}

	log.Info("Subiendo archivo a S3")
	_, err := s.Client.PutObject(ctx, input)
	if err != nil {
		log.Error("Error al subir archivo a S3", zap.Error(err))
		return errorsApp.ErrS3UploadFailed.WithMessage(fmt.Sprintf("error subiendo archivo a S3: %v", err))
	}

	log.Info("Archivo subido exitosamente")
	return nil
}

func (s *S3Storage) GetFileMetadata(ctx context.Context, key string) (*model.FileData, error) {
	log := s.log.With(
		zap.String("component", "S3Storage"),
		zap.String("method", "GetFileMetadata"),
		zap.String("key", key),
	)

	headInput := &s3.HeadObjectInput{
		Bucket: aws.String(s.Config.Storage.S3Config.BucketName),
		Key:    aws.String("audio/" + key),
	}

	log.Info("Obteniendo metadatos del archivo")
	headResult, err := s.Client.HeadObject(ctx, headInput)
	if err != nil {
		log.Error("Error al obtener metadatos del archivo", zap.Error(err))
		return nil, errorsApp.ErrS3GetMetadataFailed.WithMessage(fmt.Sprintf("error obteniendo metadata del archivo de S3: %v", err))
	}

	readableSize := formatFileSize(*headResult.ContentLength)
	log.Info("Metadatos obtenidos exitosamente", zap.String("file_size", readableSize))

	return &model.FileData{
		FilePath: "audio/" + key,
		FileType: *headResult.ContentType,
		FileSize: readableSize,
	}, nil
}

func (s *S3Storage) GetFileContent(ctx context.Context, path string, key string) (io.ReadCloser, error) {
	log := s.log.With(
		zap.String("component", "S3Storage"),
		zap.String("method", "GetFileContent"),
		zap.String("path", path),
		zap.String("key", key),
	)

	getInput := &s3.GetObjectInput{
		Bucket: aws.String(s.Config.Storage.S3Config.BucketName),
		Key:    aws.String(path + key),
	}

	log.Info("Obteniendo contenido del archivo")
	getResult, err := s.Client.GetObject(ctx, getInput)
	if err != nil {
		log.Error("Error al obtener contenido del archivo", zap.Error(err))
		return nil, errorsApp.ErrS3GetContentFailed.WithMessage(fmt.Sprintf("error obteniendo contenido del archivo %s de S3: %v", key, err))
	}

	log.Info("Contenido del archivo obtenido exitosamente")
	return getResult.Body, nil
}

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
