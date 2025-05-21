package secretmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type SecretsManager struct {
	client *secretsmanager.Client
}

func NewSecretsManager(region string) (*SecretsManager, error) {
	cfg, err := awsCfg.LoadDefaultConfig(context.TODO(), awsCfg.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("error cargando configuración AWS: %w", err)
	}

	return &SecretsManager{
		client: secretsmanager.NewFromConfig(cfg),
	}, nil
}

func (sm *SecretsManager) GetSecret(ctx context.Context, secretName string) (map[string]string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}

	result, err := sm.client.GetSecretValue(ctx, input)
	if err != nil {
		return nil, err
	}

	var secretMap map[string]string
	if err := json.Unmarshal([]byte(*result.SecretString), &secretMap); err != nil {
		return nil, fmt.Errorf("error al deserializar el secreto: %w", err)
	}

	return secretMap, nil
}
