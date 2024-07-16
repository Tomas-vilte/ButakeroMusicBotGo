package aws

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
)

type ECSClient struct {
	client *ecs.Client
}

func NewECSClient() (*ECSClient, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, err
	}

	client := ecs.NewFromConfig(cfg)

	return &ECSClient{client: client}, nil
}

func (e *ECSClient) RunTask(ctx context.Context, input *ecs.RunTaskInput) (*ecs.RunTaskOutput, error) {
	return e.client.RunTask(ctx, input)
}
