package secretmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type SecretsManager struct {
	client *secretsmanager.Client
}

func NewSecretsManager(region string) (*SecretsManager, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)

	if err != nil {
		return nil, fmt.Errorf("error al cargar la configuración de AWS: %w", err)
	}

	client := secretsmanager.NewFromConfig(cfg)

	return &SecretsManager{
		client: client,
	}, nil
}

func (sm *SecretsManager) GetSecret(ctx context.Context, secretName string) (map[string]string, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	}

	result, err := sm.client.GetSecretValue(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("error al obtener el secreto %s: %w", secretName, err)
	}

	var secretValues map[string]string
	if err := json.Unmarshal([]byte(*result.SecretString), &secretValues); err != nil {
		return nil, fmt.Errorf("error al deserializar el secreto %s: %w", secretName, err)
	}

	return secretValues, nil
}
