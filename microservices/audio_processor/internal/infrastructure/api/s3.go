package api

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	cfgAws "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func CheckS3(ctx context.Context, cfgApplication *config.Config) error {
	cfg, err := cfgAws.LoadDefaultConfig(ctx, cfgAws.WithRegion(cfgApplication.AWS.Region))
	if err != nil {
		return fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	_, err = client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(cfgApplication.Storage.S3Config.BucketName)})
	if err != nil {
		return fmt.Errorf("error en encontrar el bucket: %w", err)
	}
	return nil
}
