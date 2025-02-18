package sqs

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
	"sync"
	"testing"
	"time"
)

type MockSQSClient struct {
	mock.Mock
}

type MockLogger struct {
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

func (m *MockLogger) Info(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Debug(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) With(fields ...zapcore.Field) {
	m.Called(fields)
}

func TestNewSQSConsumer(t *testing.T) {
	// Arrange
	mockClient := new(MockSQSClient)
	mockLogger := new(MockLogger)
	config := SQSConfig{
		QueueURL:        "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
		MaxMessages:     10,
		WaitTimeSeconds: 20,
	}

	// Act
	consumer := NewSQSConsumer(mockClient, config, mockLogger)

	// Assert
	assert.NotNil(t, consumer)
	assert.Equal(t, mockClient, consumer.client)
	assert.Equal(t, config.QueueURL, consumer.queueURL)
	assert.Equal(t, config.MaxMessages, consumer.maxMessages)
	assert.Equal(t, config.WaitTimeSeconds, consumer.waitTime)
	assert.NotNil(t, consumer.messageChan)
}

func TestSQSConsumer_ConsumeMessages(t *testing.T) {
	mockClient := new(MockSQSClient)
	mockLogger := new(MockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	config := SQSConfig{
		QueueURL:        "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
		MaxMessages:     10,
		WaitTimeSeconds: 20,
	}
	consumer := NewSQSConsumer(mockClient, config, mockLogger)
	ctx, cancel := context.WithCancel(context.Background())

	err := consumer.ConsumeMessages(ctx, 0)

	assert.NoError(t, err)
	mockLogger.AssertCalled(t, "Info", "Iniciando consumo de mensajes SQS", mock.Anything)

	cancel()
	consumer.wg.Wait()
}

func TestSQSConsumer_receiveAndProcessMessages_Success(t *testing.T) {
	mockClient := new(MockSQSClient)
	mockLogger := new(MockLogger)

	ctx := context.Background()
	msgID := "test-message-id"
	receiptHandle := "test-receipt-handle"
	body := `{"status":{"status":"success"}}`

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

	config := SQSConfig{
		QueueURL:        "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
		MaxMessages:     10,
		WaitTimeSeconds: 20,
	}

	consumer := NewSQSConsumer(mockClient, config, mockLogger)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		msg := <-consumer.messageChan
		assert.Equal(t, "success", msg.Status.Status)
	}()

	consumer.receiveAndProcessMessages(ctx)

	// Wait for message processing with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Test passed
	case <-time.After(time.Second):
		t.Fatal("Test timed out waiting for message processing")
	}

	mockClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestSQSConsumer_receiveAndProcessMessages_Error(t *testing.T) {
	mockClient := new(MockSQSClient)
	mockLogger := new(MockLogger)

	ctx := context.Background()

	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	mockClient.On("ReceiveMessage", mock.Anything, mock.Anything).Return(
		&sqs.ReceiveMessageOutput{}, errors.New("test error"))

	config := SQSConfig{
		QueueURL:        "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
		MaxMessages:     10,
		WaitTimeSeconds: 20,
	}
	consumer := NewSQSConsumer(mockClient, config, mockLogger)

	// Act
	consumer.receiveAndProcessMessages(ctx)

	// Assert
	mockClient.AssertCalled(t, "ReceiveMessage", mock.Anything, mock.Anything)
	mockLogger.AssertCalled(t, "Error", "Error al recibir mensajes de la cola SQS", mock.Anything)
}

func TestSQSConsumer_handleMessage_Success(t *testing.T) {
	// Arrange
	mockClient := new(MockSQSClient)
	mockLogger := new(MockLogger)

	ctx := context.Background()
	msgID := "test-message-id"
	receiptHandle := "test-receipt-handle"
	body := `{"status":{"status":"success"}}`

	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	mockClient.On("DeleteMessage", mock.Anything, mock.Anything).Return(
		&sqs.DeleteMessageOutput{}, nil)

	config := SQSConfig{
		QueueURL:        "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
		MaxMessages:     10,
		WaitTimeSeconds: 20,
	}
	consumer := NewSQSConsumer(mockClient, config, mockLogger)

	msg := types.Message{
		MessageId:     aws.String(msgID),
		ReceiptHandle: aws.String(receiptHandle),
		Body:          aws.String(body),
	}

	go func() {
		select {
		case message := <-consumer.messageChan:
			assert.Equal(t, "success", message.Status.Status)
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
	mockLogger := new(MockLogger)

	ctx := context.Background()
	msgID := "test-message-id"
	receiptHandle := "test-receipt-handle"
	body := `{"status":{"status":"error"}}`

	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Warn", mock.Anything, mock.Anything).Return()

	mockClient.On("DeleteMessage", mock.Anything, mock.Anything).Return(
		&sqs.DeleteMessageOutput{}, nil)

	config := SQSConfig{
		QueueURL:        "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
		MaxMessages:     10,
		WaitTimeSeconds: 20,
	}
	consumer := NewSQSConsumer(mockClient, config, mockLogger)

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
	mockLogger := new(MockLogger)

	ctx := context.Background()
	msgID := "test-message-id"
	receiptHandle := "test-receipt-handle"
	body := `invalid json`

	mockLogger.On("Debug", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()

	config := SQSConfig{
		QueueURL:        "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
		MaxMessages:     10,
		WaitTimeSeconds: 20,
	}
	consumer := NewSQSConsumer(mockClient, config, mockLogger)

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
	mockLogger := new(MockLogger)

	config := SQSConfig{
		QueueURL:        "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue",
		MaxMessages:     10,
		WaitTimeSeconds: 20,
	}
	consumer := NewSQSConsumer(mockClient, config, mockLogger)

	ch := consumer.GetMessagesChannel()

	assert.NotNil(t, ch)
}
