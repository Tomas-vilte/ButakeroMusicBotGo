package sqs

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: fmt.Sprintf("http://localhost:%s", port),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(customResolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test", "test", "test",
		)),
	)

	if err != nil {
		return nil, fmt.Errorf("error cargando config AWS: %w", err)
	}

	client := sqs.NewFromConfig(cfg)

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

	cfg := SQSConfig{
		QueueURL:        container.QueueURL,
		MaxMessages:     10,
		WaitTimeSeconds: 5,
	}

	logger, err := logging.NewZapLogger()
	require.NoError(t, err)

	consumer := NewSQSConsumer(container.Client, cfg, logger)

	messageBody := `{"status":{"status":"success"}}`

	_, err = container.Client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(container.QueueURL),
		MessageBody: aws.String(messageBody),
	})
	require.NoError(t, err)

	err = consumer.ConsumeMessages(ctx, 0)
	require.NoError(t, err)

	select {
	case msg := <-consumer.messageChan:
		assert.Equal(t, "success", msg.Status.Status)
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

	cfg := SQSConfig{
		QueueURL:        container.QueueURL,
		MaxMessages:     10,
		WaitTimeSeconds: 5,
	}

	logger, err := logging.NewZapLogger()
	require.NoError(t, err)

	consumer := NewSQSConsumer(container.Client, cfg, logger)

	messageBody := `{"status":{"status":"error"}}`

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
