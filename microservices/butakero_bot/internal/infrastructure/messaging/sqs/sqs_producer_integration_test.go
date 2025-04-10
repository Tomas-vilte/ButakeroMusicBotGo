//go:build integration

package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	cfgAws "github.com/aws/aws-sdk-go-v2/config"
	"github.com/testcontainers/testcontainers-go/modules/localstack"
	"testing"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

type TestContainerProducer struct {
	Container *localstack.LocalStackContainer
	Client    *sqs.Client
	QueueURL  string
}

func setupTestContainerProducer(ctx context.Context) (*TestContainerProducer, error) {
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

	queueName := "test-song-requests-queue"
	createQueueInput := &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
	}

	result, err := client.CreateQueue(ctx, createQueueInput)
	if err != nil {
		return nil, fmt.Errorf("error creando cola: %w", err)
	}

	return &TestContainerProducer{
		Container: container,
		Client:    client,
		QueueURL:  *result.QueueUrl,
	}, nil

}

func TestProducerSQS_PublishSongRequest(t *testing.T) {
	ctx := context.Background()

	container, err := setupTestContainerProducer(ctx)
	require.NoError(t, err)
	defer func() {
		if err := container.Container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	cfg := &config.Config{
		AWS: config.AWSConfig{
			Region: "us-east-1",
		},
		QueueConfig: config.QueueConfig{
			SQSConfig: config.SQSConfig{
				Queues: &config.QueuesSQS{
					BotDownloadRequestsQueueURL: container.QueueURL,
				},
				MaxMessages:     10,
				WaitTimeSeconds: 5,
			},
		},
	}

	logger, err := logging.NewDevelopmentLogger()
	require.NoError(t, err)

	producer := &ProducerSQS{
		container.Client,
		cfg,
		logger,
	}
	require.NoError(t, err)

	testMessage := &entity.SongRequestMessage{
		InteractionID: "test-interaction-123",
		UserID:        "user-456",
		Song:          "Never Gonna Give You Up",
		ProviderType:  "youtube",
		Timestamp:     time.Now(),
	}

	err = producer.PublishSongRequest(ctx, testMessage)
	require.NoError(t, err)

	receiveResult, err := container.Client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:              aws.String(container.QueueURL),
		MaxNumberOfMessages:   1,
		WaitTimeSeconds:       5,
		MessageAttributeNames: []string{"All"},
	})
	require.NoError(t, err)
	require.Len(t, receiveResult.Messages, 1, "Expected to receive 1 message from queue")

	message := receiveResult.Messages[0]
	assert.NotNil(t, message.MessageId, "Message ID should not be nil")

	interactionID, ok := message.MessageAttributes["InteractionID"]
	require.True(t, ok, "InteractionID attribute should exist")
	assert.Equal(t, testMessage.InteractionID, *interactionID.StringValue)

	var receivedMessage entity.SongRequestMessage
	err = json.Unmarshal([]byte(*message.Body), &receivedMessage)
	require.NoError(t, err)

	assert.Equal(t, testMessage.InteractionID, receivedMessage.InteractionID)
	assert.Equal(t, testMessage.UserID, receivedMessage.UserID)
	assert.Equal(t, testMessage.Song, receivedMessage.Song)
	assert.Equal(t, testMessage.ProviderType, receivedMessage.ProviderType)

	_, err = container.Client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(container.QueueURL),
		ReceiptHandle: message.ReceiptHandle,
	})
	require.NoError(t, err)
}

func TestProducerSQS_PublishSongRequest_ContextCancelled(t *testing.T) {
	ctx := context.Background()

	container, err := setupTestContainerProducer(ctx)
	require.NoError(t, err)
	defer func() {
		if err := container.Container.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}()

	cfg := &config.Config{
		AWS: config.AWSConfig{
			Region: "us-east-1",
		},
		QueueConfig: config.QueueConfig{
			SQSConfig: config.SQSConfig{
				Queues: &config.QueuesSQS{
					BotDownloadRequestsQueueURL: container.QueueURL,
				},
			},
		},
	}

	logger, err := logging.NewDevelopmentLogger()
	require.NoError(t, err)

	producer := &ProducerSQS{
		container.Client,
		cfg,
		logger,
	}

	testMessage := &entity.SongRequestMessage{
		InteractionID: "test-interaction-123",
		UserID:        "user-456",
		Song:          "Never Gonna Give You Up",
		ProviderType:  "youtube",
		Timestamp:     time.Now(),
	}

	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	err = producer.PublishSongRequest(cancelledCtx, testMessage)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "contexto cancelado")
}
