//go:build !integration

package sqs

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
	"time"
)

type MockSQSClient struct {
	mock.Mock
}

func (m *MockSQSClient) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

func (m *MockSQSClient) ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*sqs.ReceiveMessageOutput), args.Error(1)

}

func TestNewSQSConsumer(t *testing.T) {
	// Arrange
	mockClient := new(MockSQSClient)
	mockLogger := new(logging.MockLogger)
	cfg := &config.Config{
		QueueConfig: config.QueueConfig{
			SQSConfig: config.SQSConfig{
				Queues: &config.QueuesSQS{
					BotDownloadStatusQueueURL: "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				},
				MaxMessages:     10,
				WaitTimeSeconds: 20,
			},
		},
	}

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)

	// Act
	consumer := NewSQSConsumer(mockClient, cfg, mockLogger)

	// Assert
	assert.NotNil(t, consumer)
	assert.Equal(t, mockClient, consumer.client)
	assert.Equal(t, cfg.QueueConfig.SQSConfig.Queues.BotDownloadStatusQueueURL, consumer.cfg.QueueConfig.SQSConfig.Queues.BotDownloadStatusQueueURL)
	assert.Equal(t, cfg.QueueConfig.SQSConfig.MaxMessages, consumer.cfg.QueueConfig.SQSConfig.MaxMessages)
	assert.Equal(t, cfg.QueueConfig.SQSConfig.WaitTimeSeconds, consumer.cfg.QueueConfig.SQSConfig.WaitTimeSeconds)
	assert.NotNil(t, consumer.messageChan)
}

func TestSQSConsumer_ConsumeMessages(t *testing.T) {
	mockClient := new(MockSQSClient)
	mockLogger := new(logging.MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	cfg := &config.Config{
		QueueConfig: config.QueueConfig{
			SQSConfig: config.SQSConfig{
				Queues: &config.QueuesSQS{
					BotDownloadStatusQueueURL: "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				},
				MaxMessages:     10,
				WaitTimeSeconds: 20,
			},
		},
	}
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)

	consumer := NewSQSConsumer(mockClient, cfg, mockLogger)
	ctx, cancel := context.WithCancel(context.Background())

	err := consumer.ConsumeMessages(ctx, 0)

	assert.NoError(t, err)
	mockLogger.AssertCalled(t, "Info", "Iniciando consumo de mensajes SQS", mock.Anything)

	cancel()
	consumer.wg.Wait()
}

func TestSQSConsumer_receiveAndProcessMessages_Success(t *testing.T) {
	mockClient := new(MockSQSClient)
	mockLogger := new(logging.MockLogger)

	ctx := context.Background()
	msgID := "test-message-id"
	receiptHandle := "test-receipt-handle"
	body := `{"status":"success"}`

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	mockClient.On("ReceiveMessage", ctx, mock.Anything).Return(&sqs.ReceiveMessageOutput{
		Messages: []types.Message{
			{
				MessageId:     &msgID,
				Body:          &body,
				ReceiptHandle: &receiptHandle,
			},
		},
	}, nil)
	mockClient.On("DeleteMessage", mock.Anything, mock.Anything).Return(
		&sqs.DeleteMessageOutput{}, nil)

	cfg := &config.Config{
		QueueConfig: config.QueueConfig{
			SQSConfig: config.SQSConfig{
				Queues: &config.QueuesSQS{
					BotDownloadStatusQueueURL: "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				},
				MaxMessages:     10,
				WaitTimeSeconds: 20,
			},
		},
	}

	consumer := NewSQSConsumer(mockClient, cfg, mockLogger)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		msg := <-consumer.messageChan
		assert.Equal(t, "success", msg.Status)
	}()

	consumer.receiveAndProcessMessages(ctx)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Test timed out waiting for message processing")
	}

	mockClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestSQSConsumer_receiveAndProcessMessages_Error(t *testing.T) {
	mockClient := new(MockSQSClient)
	mockLogger := new(logging.MockLogger)

	ctx := context.Background()

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockClient.On("ReceiveMessage", mock.Anything, mock.Anything).Return(
		&sqs.ReceiveMessageOutput{}, errors.New("test error"))

	cfg := &config.Config{
		QueueConfig: config.QueueConfig{
			SQSConfig: config.SQSConfig{
				Queues: &config.QueuesSQS{
					BotDownloadStatusQueueURL: "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				},
				MaxMessages:     10,
				WaitTimeSeconds: 20,
			},
		},
	}
	consumer := NewSQSConsumer(mockClient, cfg, mockLogger)

	// Act
	consumer.receiveAndProcessMessages(ctx)

	// Assert
	mockClient.AssertCalled(t, "ReceiveMessage", mock.Anything, mock.Anything)
	mockLogger.AssertCalled(t, "Error", "Error al recibir mensajes de la cola SQS", mock.Anything)
}

func TestSQSConsumer_handleMessage_Success(t *testing.T) {
	// Arrange
	mockClient := new(MockSQSClient)
	mockLogger := new(logging.MockLogger)

	ctx := context.Background()
	msgID := "test-message-id"
	receiptHandle := "test-receipt-handle"
	body := `{"status":"success"}`

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	mockClient.On("DeleteMessage", mock.Anything, mock.Anything).Return(
		&sqs.DeleteMessageOutput{}, nil)

	cfg := &config.Config{
		QueueConfig: config.QueueConfig{
			SQSConfig: config.SQSConfig{
				Queues: &config.QueuesSQS{
					BotDownloadStatusQueueURL: "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				},
				MaxMessages:     10,
				WaitTimeSeconds: 20,
			},
		},
	}
	consumer := NewSQSConsumer(mockClient, cfg, mockLogger)

	msg := types.Message{
		MessageId:     aws.String(msgID),
		ReceiptHandle: aws.String(receiptHandle),
		Body:          aws.String(body),
	}

	go func() {
		select {
		case message := <-consumer.messageChan:
			assert.Equal(t, "success", message.Status)
		case <-time.After(100 * time.Millisecond):
			t.Error("Timeout waiting for message in channel")
		}
	}()

	// Act
	consumer.handleMessage(ctx, msg)

	// Assert
	mockClient.AssertCalled(t, "DeleteMessage", mock.Anything, mock.Anything)
	mockLogger.AssertCalled(t, "Debug", "Mensaje recibido", mock.Anything)
	mockLogger.AssertCalled(t, "Info", "Mensaje procesado exitosamente", mock.Anything)
}

func TestSQSConsumer_handleMessage_WarningStatus(t *testing.T) {
	// Arrange
	mockClient := new(MockSQSClient)
	mockLogger := new(logging.MockLogger)

	ctx := context.Background()
	msgID := "test-message-id"
	receiptHandle := "test-receipt-handle"
	body := `{"status":"error"}`

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

	mockClient.On("DeleteMessage", mock.Anything, mock.Anything).Return(
		&sqs.DeleteMessageOutput{}, nil)

	cfg := &config.Config{
		QueueConfig: config.QueueConfig{
			SQSConfig: config.SQSConfig{
				Queues: &config.QueuesSQS{
					BotDownloadStatusQueueURL: "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				},
				MaxMessages:     10,
				WaitTimeSeconds: 20,
			},
		},
	}
	consumer := NewSQSConsumer(mockClient, cfg, mockLogger)

	msg := types.Message{
		MessageId:     aws.String(msgID),
		ReceiptHandle: aws.String(receiptHandle),
		Body:          aws.String(body),
	}

	// Act
	consumer.handleMessage(ctx, msg)

	// Assert
	mockClient.AssertCalled(t, "DeleteMessage", mock.Anything, mock.Anything)
	mockLogger.AssertCalled(t, "Debug", "Mensaje recibido", mock.Anything)
	mockLogger.AssertCalled(t, "Warn", "Mensaje recibido con estado de error", mock.Anything)
}

func TestSQSConsumer_handleMessage_UnmarshalError(t *testing.T) {
	// Arrange
	mockClient := new(MockSQSClient)
	mockLogger := new(logging.MockLogger)

	ctx := context.Background()
	msgID := "test-message-id"
	receiptHandle := "test-receipt-handle"
	body := `invalid json`

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	cfg := &config.Config{
		QueueConfig: config.QueueConfig{
			SQSConfig: config.SQSConfig{
				Queues: &config.QueuesSQS{
					BotDownloadStatusQueueURL: "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				},
				MaxMessages:     10,
				WaitTimeSeconds: 20,
			},
		},
	}
	consumer := NewSQSConsumer(mockClient, cfg, mockLogger)

	msg := types.Message{
		MessageId:     aws.String(msgID),
		ReceiptHandle: aws.String(receiptHandle),
		Body:          aws.String(body),
	}

	// Act
	consumer.handleMessage(ctx, msg)

	// Assert
	mockLogger.AssertCalled(t, "Debug", "Mensaje recibido", mock.Anything)
	mockLogger.AssertCalled(t, "Error", "Error al deserializar mensaje", mock.Anything)
}

func TestSQSConsumer_GetMessagesChannel(t *testing.T) {
	mockClient := new(MockSQSClient)
	mockLogger := new(logging.MockLogger)

	cfg := &config.Config{
		QueueConfig: config.QueueConfig{
			SQSConfig: config.SQSConfig{
				Queues: &config.QueuesSQS{
					BotDownloadStatusQueueURL: "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
				},
				MaxMessages:     10,
				WaitTimeSeconds: 20,
			},
		},
	}
	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)

	consumer := NewSQSConsumer(mockClient, cfg, mockLogger)

	ch := consumer.GetMessagesChannel()

	assert.NotNil(t, ch)
}
