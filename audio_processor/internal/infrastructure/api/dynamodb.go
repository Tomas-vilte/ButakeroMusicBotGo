package api

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	cfgAws "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func CheckDynamoDB(ctx context.Context, cfgApplication *config.Config) (*DynamoDBMetadata, error) {
	cfg, err := cfgAws.LoadDefaultConfig(ctx, cfgAws.WithRegion(cfgApplication.AWS.Region))
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	_, err = client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: aws.String(cfgApplication.Database.DynamoDB.Tables.Songs)})
	if err != nil {
		return nil, fmt.Errorf("error al obtener info de la tabla: %w", err)
	}
	return &DynamoDBMetadata{
		TableName: cfgApplication.Database.DynamoDB.Tables.Songs,
	}, nil
}
