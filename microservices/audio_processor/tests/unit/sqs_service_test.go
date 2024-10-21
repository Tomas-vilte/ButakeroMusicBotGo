package unit

import (
	"context"
	"encoding/json"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue"
	sqsService "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/queue/sqs"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestSendMessage(t *testing.T) {
	t.Run("NewSQSService", func(t *testing.T) {
		mockLogger := new(MockLogger)
		cfg := config.Config{
			Region:    "us-east-1",
			AccessKey: "test-access-key",
			SecretKey: "test-secret-key",
			QueueURL:  "test-queue-url",
		}

		service, err := sqsService.NewSQSService(cfg, mockLogger)

		assert.NoError(t, err)
		assert.NotNil(t, service)
		assert.Equal(t, cfg, service.Config)
	})

	t.Run("SendMessage", func(t *testing.T) {
		mockClient := new(MockSQSClient)
		mockLogger := new(MockLogger)

		cfg := config.Config{
			QueueURL: "test-queue-url",
		}

		service := &sqsService.SQSService{
			Config: cfg,
			Client: mockClient,
			Log:    mockLogger,
		}

		message := queue.Message{
			ID:      "test-message-id",
			Content: "test-content",
		}

		expectedBody, _ := json.Marshal(message)
		expectedInput := &sqs.SendMessageInput{
			QueueUrl:    aws.String(cfg.QueueURL),
			MessageBody: aws.String(string(expectedBody)),
		}

		mockClient.On("SendMessage", mock.Anything, expectedInput, mock.Anything).Return(&sqs.SendMessageOutput{}, nil)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()

		err := service.SendMessage(context.Background(), message)

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("SendMessageError", func(t *testing.T) {
		mockClient := new(MockSQSClient)
		mockLogger := new(MockLogger)
		cfg := config.Config{
			QueueURL: "test-queue-url",
		}

		service := &sqsService.SQSService{
			Config: cfg,
			Client: mockClient,
			Log:    mockLogger,
		}

		message := queue.Message{
			ID:      "test-message-id",
			Content: "test-content",
		}

		expectedBody, _ := json.Marshal(message)
		expectedInput := &sqs.SendMessageInput{
			QueueUrl:    aws.String(cfg.QueueURL),
			MessageBody: aws.String(string(expectedBody)),
		}

		mockClient.On("SendMessage", mock.Anything, expectedInput, mock.Anything).
			Return(&sqs.SendMessageOutput{}, errors.New("SQS error")).Times(3)
		mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

		err := service.SendMessage(context.Background(), message)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error al enviar mensaje a SQS después de varios intentos")
		mockClient.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

}

func TestReceiveMessage(t *testing.T) {
	mockClient := new(MockSQSClient)
	mockLogger := new(MockLogger)

	service := &sqsService.SQSService{
		Config: config.Config{
			QueueURL: "test-queue-url",
		},
		Client: mockClient,
		Log:    mockLogger,
	}

	t.Run("Success", func(t *testing.T) {
		messageBody := queue.MessageBody{
			ID:      "test-id",
			Content: "test content",
		}
		jsonBody, _ := json.Marshal(messageBody)
		expectedOutput := &sqs.ReceiveMessageOutput{
			Messages: []types.Message{
				{
					Body:          aws.String(string(jsonBody)),
					ReceiptHandle: aws.String("test-receipt-handle"),
				},
			},
		}
		mockClient.On("ReceiveMessage", mock.Anything, mock.Anything, mock.Anything).Return(expectedOutput, nil).Once()
		mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()

		messages, err := service.ReceiveMessage(context.Background())

		assert.NoError(t, err)
		assert.NotNil(t, messages)
		assert.Len(t, messages, 1)
		assert.Equal(t, messageBody.ID, messages[0].ID)
		assert.Equal(t, messageBody.Content, messages[0].Content)
		assert.Equal(t, "test-receipt-handle", messages[0].ReceiptHandle)

		mockClient.AssertExpectations(t)
	})

	t.Run("No Messages", func(t *testing.T) {
		expectedOutput := &sqs.ReceiveMessageOutput{
			Messages: []types.Message{},
		}

		mockClient.On("ReceiveMessage", mock.Anything, mock.Anything, mock.Anything).Return(expectedOutput, nil).Once()
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()

		messages, err := service.ReceiveMessage(context.Background())

		assert.NoError(t, err)
		assert.Empty(t, messages)
		mockClient.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockClient.On("ReceiveMessage", mock.Anything, mock.Anything, mock.Anything).Return(&sqs.ReceiveMessageOutput{}, assert.AnError).Once()
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()

		messages, err := service.ReceiveMessage(context.Background())

		assert.Error(t, err)
		assert.Nil(t, messages)
		mockClient.AssertExpectations(t)
	})
}

func TestDeleteMessage(t *testing.T) {
	mockClient := new(MockSQSClient)
	mockLogger := new(MockLogger)
	service := &sqsService.SQSService{
		Config: config.Config{
			QueueURL: "test-queue-url",
		},
		Log:    mockLogger,
		Client: mockClient,
	}

	t.Run("Success", func(t *testing.T) {
		mockClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything).Return(&sqs.DeleteMessageOutput{}, nil).Once()
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()

		err := service.DeleteMessage(context.Background(), "receipt-handle")

		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		mockClient.On("DeleteMessage", mock.Anything, mock.Anything, mock.Anything).Return(&sqs.DeleteMessageOutput{}, assert.AnError).Once()
		mockLogger.On("Error", mock.Anything, mock.Anything).Return()

		err := service.DeleteMessage(context.Background(), "receipt-handle")

		assert.Error(t, err)
		mockClient.AssertExpectations(t)
	})
}
