////go:build integration

package sqs

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	cfgAws "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	"testing"
	"time"
)

type TestContainer struct {
	Container *localstack.LocalStackContainer
	Client    *sqs.Client
	QueueURL  string
}

func setupTestContainer(ctx context.Context) (*TestContainer, error) {
	container, err := localstack.Run(ctx, "localstack/localstack:4.1.1", testcontainers.WithEnv(map[string]string{
		"SERVICES":           "sqs",
		"AWS_DEFAULT_REGION": "us-east-1",
		"EDGE_PORT":          "4566",
	}))
	if err != nil {
		return nil, fmt.Errorf("error iniciando contenedor: %w", err)
	}

	port, err := container.MappedPort(ctx, "4566")
	if err != nil {
		return nil, fmt.Errorf("error obteniendo puerto mapeado: %w", err)
	}

	customEndpoint := fmt.Sprintf("http://localhost:%s", port)
	cfg, err := cfgAws.LoadDefaultConfig(ctx,
		cfgAws.WithRegion("us-east-1"),
		cfgAws.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "test",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("error cargando config AWS: %w", err)
	}

	client := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(customEndpoint)
	})

	queueName := "test-queue"
	createQueueInput := &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	}

	result, err := client.CreateQueue(ctx, createQueueInput)
	if err != nil {
		return nil, fmt.Errorf("error creando cola: %w", err)
	}

	return &TestContainer{
		Container: container,
		Client:    client,
		QueueURL:  *result.QueueUrl,
	}, nil

}

func TestSQSConsumerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := context.Background()
	container, err := setupTestContainer(ctx)
	require.NoError(t, err)

	defer func() {
		if err := container.Container.Terminate(ctx); err != nil {
			t.Fatalf("error terminando contenedor: %v", err)
		}
	}()

	cfg := &config.Config{
		QueueConfig: config.QueueConfig{
			SQSConfig: config.SQSConfig{
				QueueURL:        container.QueueURL,
				MaxMessages:     10,
				WaitTimeSeconds: 5,
			},
		},
	}

	logger, err := logging.NewDevelopmentLogger()
	require.NoError(t, err)

	consumer := NewSQSConsumer(container.Client, cfg, logger)

	messageBody := `{"status":"success"}`

	_, err = container.Client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(container.QueueURL),
		MessageBody: aws.String(messageBody),
	})
	require.NoError(t, err)

	err = consumer.ConsumeMessages(ctx, 0)
	require.NoError(t, err)

	select {
	case msg := <-consumer.messageChan:
		assert.Equal(t, "success", msg.Status)
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for message")
	}
}

func TestSQSConsumerIntegration_ErrorStatus(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	ctx := context.Background()
	container, err := setupTestContainer(ctx)
	require.NoError(t, err)

	defer func() {
		if err := container.Container.Terminate(ctx); err != nil {
			t.Fatalf("error terminando contenedor: %v", err)
		}
	}()

	cfg := &config.Config{
		QueueConfig: config.QueueConfig{
			SQSConfig: config.SQSConfig{
				QueueURL:        container.QueueURL,
				MaxMessages:     10,
				WaitTimeSeconds: 5,
			},
		},
	}

	logger, err := logging.NewDevelopmentLogger()
	require.NoError(t, err)

	consumer := NewSQSConsumer(container.Client, cfg, logger)

	messageBody := `{"status":"error"}`

	_, err = container.Client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(container.QueueURL),
		MessageBody: aws.String(messageBody),
	})
	require.NoError(t, err)

	err = consumer.ConsumeMessages(ctx, 0)
	require.NoError(t, err)

	select {
	case <-consumer.messageChan:
		t.Fatal("no debería recibir mensajes con estado de error")
	case <-time.After(2 * time.Second):
	}
}
