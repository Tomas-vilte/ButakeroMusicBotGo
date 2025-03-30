//go:build integration

package sqs

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"testing"
	"time"
)

const (
	localstackImage = "localstack/localstack:latest"
	sqsPort         = "4566/tcp"
	queueName       = "test-queue"
	region          = "us-east-1"
	accessKey       = "test"
	secretKey       = "test"
)

type LocalstackContainer struct {
	testcontainers.Container
	URI string
}

// setupLocalstack inicia un contenedor de Localstack
func setupLocalstack(ctx context.Context) (*LocalstackContainer, error) {
	natSQSPort := nat.Port(sqsPort)
	req := testcontainers.ContainerRequest{
		Image:        localstackImage,
		ExposedPorts: []string{sqsPort},
		Env: map[string]string{
			"SERVICES":              "sqs",
			"DEBUG":                 "1",
			"DATA_DIR":              "/tmp/localstack/data",
			"DOCKER_HOST":           "unix:///var/run/docker.sock",
			"AWS_DEFAULT_REGION":    region,
			"AWS_ACCESS_KEY_ID":     accessKey,
			"AWS_SECRET_ACCESS_KEY": secretKey,
		},
		WaitingFor: wait.ForListeningPort(natSQSPort).WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, natSQSPort)
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("http://%s:%s", hostIP, mappedPort.Port())

	return &LocalstackContainer{
		Container: container,
		URI:       uri,
	}, nil
}

// createSQSClient crea un cliente de SQS que apunta al contenedor de Localstack
func createSQSClient(ctx context.Context, endpoint string) (*sqs.Client, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           endpoint,
			SigningRegion: region,
		}, nil
	})

	cfg, err := awsConfig.LoadDefaultConfig(ctx,
		awsConfig.WithRegion(region),
		awsConfig.WithEndpointResolverWithOptions(customResolver),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, err
	}

	return sqs.NewFromConfig(cfg), nil
}

// createQueue crea una cola SQS en el contenedor localstack
func createQueue(ctx context.Context, client *sqs.Client) (string, error) {
	resp, err := client.CreateQueue(ctx, &sqs.CreateQueueInput{
		QueueName: aws.String(queueName),
		Attributes: map[string]string{
			"MessageRetentionPeriod": "86400",
		},
	})
	if err != nil {
		return "", err
	}
	return *resp.QueueUrl, nil
}

func TestSQSServiceWithTestcontainers(t *testing.T) {
	if testing.Short() {
		t.Skip("Saltando test de integración en modo corto")
	}

	ctx := context.Background()

	// Iniciar el contenedor localstack
	localstack, err := setupLocalstack(ctx)
	require.NoError(t, err)
	defer func() {
		if err := localstack.Terminate(ctx); err != nil {
			t.Logf("Error al terminar el contenedor: %v", err)
		}
	}()

	// Crear un cliente SQS que apunte al contenedor
	sqsClient, err := createSQSClient(ctx, localstack.URI)
	require.NoError(t, err)

	// Crear la cola de SQS
	queueURL, err := createQueue(ctx, sqsClient)
	require.NoError(t, err)

	// Configurar el servicio SQS con el endpoint del contenedor
	cfg := &config.Config{
		AWS: config.AWSConfig{
			Region: region,
		},
		Messaging: config.MessagingConfig{
			SQS: &config.SQSConfig{
				QueueURL: queueURL,
			},
		},
	}

	log, err := logger.NewProductionLogger()
	require.NoError(t, err)

	// Crear el servicio SQS manualmente con el cliente que apunta al contenedor
	service := &SQSService{
		Client: sqsClient,
		Config: cfg,
		Log:    log,
	}

	// Ejecutar las pruebas
	t.Run("End-to-End Message Flow", func(t *testing.T) {
		// Crear y enviar un mensaje
		message := &model.MediaProcessingMessage{
			VideoID: "test_video_id",
			FileData: &model.FileData{
				FilePath: "/path/to/test/file",
				FileSize: "1024",
				FileType: "mp3",
			},
			PlatformMetadata: &model.PlatformMetadata{
				Title:      "Test Title",
				DurationMs: 3234,
			},
		}

		err := service.SendMessage(ctx, message)
		assert.NoError(t, err)

		// Recibir el mensaje
		receivedMessages, err := service.ReceiveMessage(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, receivedMessages)
		assert.Equal(t, message.VideoID, receivedMessages[0].VideoID)
		assert.Equal(t, message.FileData.FilePath, receivedMessages[0].FileData.FilePath)

		// Eliminar el mensaje
		err = service.DeleteMessage(ctx, receivedMessages[0].ReceiptHandle)
		assert.NoError(t, err)

		// Verificar que no hay más mensajes
		time.Sleep(1 * time.Second) // Breve pausa para asegurar que SQS procese la eliminación
		messages, err := service.ReceiveMessage(ctx)
		assert.NoError(t, err)
		assert.Empty(t, messages)
	})

	t.Run("Concurrent Message Processing", func(t *testing.T) {
		numMessages := 5
		messages := make([]*model.MediaProcessingMessage, numMessages)
		for i := 0; i < numMessages; i++ {
			messages[i] = &model.MediaProcessingMessage{
				VideoID: fmt.Sprintf("video_id_%d", i),
				FileData: &model.FileData{
					FilePath: fmt.Sprintf("/path/test_%d", i),
					FileSize: "1234",
					FileType: "dca",
				},
				PlatformMetadata: &model.PlatformMetadata{},
			}
		}

		errChan := make(chan error, numMessages)
		for _, msg := range messages {
			go func(m *model.MediaProcessingMessage) {
				errChan <- service.SendMessage(ctx, m)
			}(msg)
		}

		for i := 0; i < numMessages; i++ {
			assert.NoError(t, <-errChan)
		}

		// Recibir todos los mensajes
		var allMessages []model.MediaProcessingMessage
		for len(allMessages) < numMessages {
			receivedMessages, err := service.ReceiveMessage(ctx)
			assert.NoError(t, err)
			if len(receivedMessages) > 0 {
				allMessages = append(allMessages, receivedMessages...)
				// Eliminar los mensajes recibidos
				for _, msg := range receivedMessages {
					err := service.DeleteMessage(ctx, msg.ReceiptHandle)
					assert.NoError(t, err)
				}
			}
		}

		assert.Equal(t, numMessages, len(allMessages))
	})
}
