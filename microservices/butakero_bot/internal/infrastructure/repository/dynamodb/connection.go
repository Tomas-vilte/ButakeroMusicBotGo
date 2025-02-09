package dynamodb

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type ClientOptions func(*dynamodb.Options)

// NewDynamoDBClient crea un cliente de DynamoDB configurado
func NewDynamoDBClient(ctx context.Context, region string, opts ...ClientOptions) (*dynamodb.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	client := dynamodb.NewFromConfig(cfg, func(options *dynamodb.Options) {
		for _, opt := range opts {
			opt(options)
		}
	})
	return client, nil
}
