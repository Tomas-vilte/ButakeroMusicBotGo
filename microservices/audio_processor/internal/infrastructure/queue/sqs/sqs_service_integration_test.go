//go:build integration

package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	awsSqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"
)

const (
	localstackImage = "localstack/localstack:latest"
	queueName       = "test-queue"
	region          = "us-east-1"
)

type TestingSuite struct {
	t             *testing.T
	container     testcontainers.Container
	cfg           *config.Config
	logger        logger.Logger
	sqsClient     *awsSqs.Client
	sqsEndpoint   string
	queueURL      string
	producerSQS   *ProducerSQS
	consumerSQS   *ConsumerSQS
	ctx           context.Context
	cancelContext context.CancelFunc
}

func setupTestingSuite(t *testing.T) *TestingSuite {
	ctx, cancel := context.WithCancel(context.Background())

	log, err := logger.NewDevelopmentLogger()
	require.NoError(t, err)

	suite := &TestingSuite{
		t:             t,
		logger:        log,
		ctx:           ctx,
		cancelContext: cancel,
	}

	suite.setupLocalstack()

	suite.setupSQSClient()
	suite.createQueue()

	suite.cfg = &config.Config{
		AWS: config.AWSConfig{
			Region: region,
		},
		Messaging: config.MessagingConfig{
			SQS: &config.SQSConfig{
				QueueURL: suite.queueURL,
			},
		},
	}

	suite.setupProducerAndConsumer()

	return suite
}

func (suite *TestingSuite) setupLocalstack() {

	// Crear y arrancar contenedor de localstack
	req := testcontainers.ContainerRequest{
		Image:        localstackImage,
		ExposedPorts: []string{"4566/tcp"},
		Env: map[string]string{
			"SERVICES":              "sqs",
			"DEFAULT_REGION":        region,
			"AWS_ACCESS_KEY_ID":     "test",
			"AWS_SECRET_ACCESS_KEY": "test",
		},
		WaitingFor: wait.ForListeningPort("4566/tcp"),
	}

	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(suite.t, err, "Error al crear container de localstack")

	mappedPort, err := container.MappedPort(suite.ctx, "4566")
	require.NoError(suite.t, err, "Error al obtener puerto mapeado")

	host, err := container.Host(suite.ctx)
	require.NoError(suite.t, err, "Error al obtener host")

	suite.container = container
	suite.sqsEndpoint = fmt.Sprintf("http://%s:%s", host, mappedPort.Port())

	suite.logger.Info("Localstack container iniciado",
		zap.String("endpoint", suite.sqsEndpoint))
}

func (suite *TestingSuite) setupSQSClient() {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           suite.sqsEndpoint,
			SigningRegion: region,
		}, nil
	})

	cfg, err := awsCfg.LoadDefaultConfig(suite.ctx,
		awsCfg.WithRegion(region),
		awsCfg.WithEndpointResolverWithOptions(customResolver),
		awsCfg.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "test",
				SecretAccessKey: "test",
			}, nil
		})),
	)
	require.NoError(suite.t, err, "Error al cargar configuración de AWS")

	suite.sqsClient = awsSqs.NewFromConfig(cfg)
}

func (suite *TestingSuite) createQueue() {
	createQueueInput := &awsSqs.CreateQueueInput{
		QueueName: aws.String(queueName),
		Attributes: map[string]string{
			"MessageRetentionPeriod": "86400",
		},
	}

	createQueueOutput, err := suite.sqsClient.CreateQueue(suite.ctx, createQueueInput)
	require.NoError(suite.t, err, "Error al crear cola SQS")

	suite.queueURL = *createQueueOutput.QueueUrl
	suite.logger.Info("Cola SQS creada", zap.String("queue_url", suite.queueURL))
}

func (suite *TestingSuite) setupProducerAndConsumer() {
	producer := &ProducerSQS{
		suite.sqsClient,
		suite.logger,
		suite.cfg,
	}
	suite.producerSQS = producer

	consumer := &ConsumerSQS{
		suite.cfg,
		suite.logger,
		suite.sqsClient,
	}
	suite.consumerSQS = consumer
}

func (suite *TestingSuite) tearDown() {
	suite.cancelContext()
	if suite.container != nil {
		err := suite.container.Terminate(context.Background())
		assert.NoError(suite.t, err, "Error al terminar container")
	}
}

func (suite *TestingSuite) sendMessageToQueue(msg *model.MediaRequest) {
	body, err := json.Marshal(msg)
	require.NoError(suite.t, err, "Error al serializar mensaje")

	input := &awsSqs.SendMessageInput{
		MessageBody: aws.String(string(body)),
		QueueUrl:    aws.String(suite.queueURL),
	}

	_, err = suite.sqsClient.SendMessage(suite.ctx, input)
	require.NoError(suite.t, err, "Error al enviar mensaje a SQS")
}

func (suite *TestingSuite) receiveMessagesFromQueue() []string {
	input := &awsSqs.ReceiveMessageInput{
		QueueUrl:            aws.String(suite.queueURL),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     1,
	}

	result, err := suite.sqsClient.ReceiveMessage(suite.ctx, input)
	require.NoError(suite.t, err, "Error al recibir mensajes de SQS")

	var messages []string
	for _, msg := range result.Messages {
		messages = append(messages, *msg.Body)
	}

	return messages
}

func TestProducerSQS_Publish(t *testing.T) {
	suite := setupTestingSuite(t)
	defer suite.tearDown()

	testMsg := &model.MediaProcessingMessage{
		VideoID: "test-video-id",
		FileData: &model.FileData{
			FilePath: "test-file.mp3",
			FileSize: "1024",
		},
		PlatformMetadata: &model.PlatformMetadata{
			Platform: "youtube",
		},
		Message: "Test message",
		Success: true,
		Status:  "PROCESSING",
	}

	err := suite.producerSQS.Publish(suite.ctx, testMsg)

	assert.NoError(t, err, "Error al publicar mensaje")

	messages := suite.receiveMessagesFromQueue()
	assert.Len(t, messages, 1, "Se esperaba 1 mensaje en la cola")

	var receivedMsg model.MediaProcessingMessage
	err = json.Unmarshal([]byte(messages[0]), &receivedMsg)
	assert.NoError(t, err, "Error al deserializar mensaje")

	assert.Equal(t, testMsg.VideoID, receivedMsg.VideoID)
	assert.Equal(t, testMsg.Message, receivedMsg.Message)
	assert.Equal(t, testMsg.Status, receivedMsg.Status)
	assert.Equal(t, testMsg.Success, receivedMsg.Success)
	assert.Equal(t, testMsg.FileData.FilePath, receivedMsg.FileData.FilePath)
	assert.Equal(t, testMsg.FileData.FileSize, receivedMsg.FileData.FileSize)
	assert.Equal(t, testMsg.PlatformMetadata.Platform, receivedMsg.PlatformMetadata.Platform)
}

func TestConsumerSQS_GetRequestsChannel(t *testing.T) {
	suite := setupTestingSuite(t)
	defer suite.tearDown()

	testMsg := &model.MediaRequest{
		InteractionID: "test-interaction-id",
		UserID:        "test-user-id",
		Song:          "test-song",
		ProviderType:  "youtube",
		Timestamp:     time.Now(),
	}

	suite.sendMessageToQueue(testMsg)

	msgChan, err := suite.consumerSQS.GetRequestsChannel(suite.ctx)
	assert.NoError(t, err, "Error al obtener canal de mensajes")

	select {
	case receivedMsg := <-msgChan:
		assert.NotNil(t, receivedMsg, "Mensaje recibido no debe ser nil")
		assert.Equal(t, testMsg.InteractionID, receivedMsg.InteractionID)
		assert.Equal(t, testMsg.UserID, receivedMsg.UserID)
		assert.Equal(t, testMsg.Song, receivedMsg.Song)
		assert.Equal(t, testMsg.ProviderType, receivedMsg.ProviderType)
	case <-time.After(5 * time.Second):
		assert.Fail(t, "Timeout esperando mensaje del canal")
	}
}

func TestProducerConsumerIntegration(t *testing.T) {
	suite := setupTestingSuite(t)
	defer suite.tearDown()

	testMsg := &model.MediaProcessingMessage{
		VideoID: "integration-test-video-id",
		FileData: &model.FileData{
			FilePath: "integration-test.mp3",
			FileSize: "2048",
		},
		PlatformMetadata: &model.PlatformMetadata{
			Platform: "youtube",
		},
		Message: "Integration test message",
		Success: true,
		Status:  "COMPLETED",
	}

	mediaRequest := &model.MediaRequest{
		InteractionID: "integration-test-interaction-id",
		UserID:        "integration-test-user",
		Song:          "integration-test-song",
		ProviderType:  "youtube",
		Timestamp:     time.Now(),
	}

	err := suite.producerSQS.Publish(suite.ctx, testMsg)
	assert.NoError(t, err, "Error al publicar mensaje con productor")

	messages := suite.receiveMessagesFromQueue()
	assert.Len(t, messages, 1, "Se esperaba 1 mensaje en la cola")

	suite.sendMessageToQueue(mediaRequest)

	msgChan, err := suite.consumerSQS.GetRequestsChannel(suite.ctx)
	assert.NoError(t, err, "Error al obtener canal de mensajes")

	select {
	case receivedMsg := <-msgChan:
		assert.NotNil(t, receivedMsg, "Mensaje recibido no debe ser nil")
		assert.Equal(t, mediaRequest.InteractionID, receivedMsg.InteractionID)
		assert.Equal(t, mediaRequest.UserID, receivedMsg.UserID)
		assert.Equal(t, mediaRequest.Song, receivedMsg.Song)
		assert.Equal(t, mediaRequest.ProviderType, receivedMsg.ProviderType)
	case <-time.After(5 * time.Second):
		assert.Fail(t, "Timeout esperando mensaje del canal")
	}
}

func TestLongPollConsumer(t *testing.T) {
	suite := setupTestingSuite(t)
	defer suite.tearDown()

	msgChan, err := suite.consumerSQS.GetRequestsChannel(suite.ctx)
	assert.NoError(t, err, "Error al obtener canal de mensajes")

	delayedMsg := &model.MediaRequest{
		InteractionID: "delayed-test-id",
		UserID:        "delayed-user",
		Song:          "delayed-song",
		ProviderType:  "delayed-provider",
		Timestamp:     time.Now(),
	}

	go func() {
		time.Sleep(2 * time.Second)
		suite.sendMessageToQueue(delayedMsg)
	}()

	select {
	case receivedMsg := <-msgChan:
		assert.NotNil(t, receivedMsg, "Mensaje recibido no debe ser nil")
		assert.Equal(t, delayedMsg.InteractionID, receivedMsg.InteractionID)
		assert.Equal(t, delayedMsg.UserID, receivedMsg.UserID)
	case <-time.After(10 * time.Second):
		assert.Fail(t, "Timeout esperando mensaje del canal (long polling)")
	}
}
