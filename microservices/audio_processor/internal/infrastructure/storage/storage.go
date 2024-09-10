package storage

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"io"
)

type (
	Storage interface {
		UploadFile(ctx context.Context, key string, body io.Reader) error
	}

	S3Client interface {
		PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	}

	S3Storage struct {
		Client     S3Client
		BucketName string
	}
)

func NewS3Storage(bucketName string, region string) (*S3Storage, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &S3Storage{
		Client:     client,
		BucketName: bucketName,
	}, nil
}

func (s *S3Storage) UploadFile(ctx context.Context, key string, body io.Reader) error {
	if body == nil {
		return fmt.Errorf("el cuerpo no puede ser nulo")
	}
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String("audio/" + key),
		Body:   body,
	}

	_, err := s.Client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("error subiendo archivo a S3: %w", err)
	}
	return nil
}
