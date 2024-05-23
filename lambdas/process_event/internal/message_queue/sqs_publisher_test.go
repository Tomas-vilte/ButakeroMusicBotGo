package message_queue

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"testing"
)

type MockSQSClient struct {
	mock.Mock
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockSQSClient) SendMessageWithContext(ctx aws.Context, input *sqs.SendMessageInput, opts ...request.Option) (*sqs.SendMessageOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*sqs.SendMessageOutput), args.Error(1)
}

func TestSQSPublisher_Publish_Success(t *testing.T) {
	mockClient := &MockSQSClient{}
	mockLogger := &MockLogger{}

	event := common.Event{
		Action: "test_action",
		Release: common.Release{
			TagName: "v1.0.0",
		},
	}
	eventJSON, _ := json.Marshal(event)

	mockClient.On("SendMessageWithContext", mock.Anything, &sqs.SendMessageInput{
		QueueUrl:    aws.String("test_queue_url"),
		MessageBody: aws.String(string(eventJSON)),
	}).Return(&sqs.SendMessageOutput{}, nil)
	mockLogger.On("Info", "Mensaje enviado a SQS con exito", []zap.Field{
		zap.String("QueueURL", "test_queue_url"),
	}).Return()

	publisher := NewSQSPublisher(mockClient, map[string]string{"test_action": "test_queue_url"}, mockLogger)
	err := publisher.Publish(context.Background(), event, "test_action")

	assert.Nil(t, err)
	mockClient.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestSQSPublisher_Publish_SendMessageError(t *testing.T) {
	mockClient := &MockSQSClient{}
	mockLogger := &MockLogger{}

	event := common.Event{
		Action: "test_action",
		Release: common.Release{
			TagName: "v1.0.0",
		},
	}

	mockClient.On("SendMessageWithContext", mock.Anything, mock.Anything).Return(&sqs.SendMessageOutput{}, errors.New("hubo un error en enviar"))
	mockLogger.On("Error", "Error enviando el mensaje a SQS", mock.Anything).Return()

	publisher := NewSQSPublisher(mockClient, map[string]string{"test_action": "test_queue_url"}, mockLogger)
	err := publisher.Publish(context.Background(), event, "test_action")

	assert.NotNil(t, err)
	mockLogger.AssertExpectations(t)
}

func TestSQSPublisher_Publish_SerializationError(t *testing.T) {
	mockClient := &MockSQSClient{}
	mockLogger := &MockLogger{}

	eventFail := map[string]interface{}{
		"foo": make(chan int),
	}

	mockLogger.On("Error", "Error serializando el evento", mock.Anything).Return()

	publisher := NewSQSPublisher(mockClient, map[string]string{"test_action": "test_queue_url"}, mockLogger)
	err := publisher.Publish(context.Background(), eventFail, "test_action")

	assert.NotNil(t, err)
	mockClient.AssertNotCalled(t, "SendMessageWithContext", mock.Anything, mock.Anything)
	mockLogger.AssertExpectations(t)
}

func TestSQSPublisher_Publish_QueueURLNotFound(t *testing.T) {
	mockClient := &MockSQSClient{}
	mockLogger := &MockLogger{}

	event := common.Event{
		Action: "test_action",
		Release: common.Release{
			TagName: "v1.0.0",
		},
	}

	publisher := NewSQSPublisher(mockClient, map[string]string{}, mockLogger)
	mockLogger.On("Error", "No se encontró URL de cola para el tipo de evento", mock.Anything).Return()
	err := publisher.Publish(context.Background(), event, "test_action")

	assert.NotNil(t, err)
	assert.Equal(t, "URL de cola no encontrada para el tipo de evento: test_action", err.Error())
	mockLogger.AssertExpectations(t)
}
