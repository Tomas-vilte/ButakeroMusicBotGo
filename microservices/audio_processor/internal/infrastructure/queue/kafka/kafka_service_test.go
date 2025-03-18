//go:build !integration

package kafka

import (
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestKafkaService(t *testing.T) {
	cfg := &config.Config{
		Messaging: config.MessagingConfig{
			Kafka: &config.KafkaConfig{
				Brokers: []string{"localhost:9092"},
				Topic:   "test-topic",
			},
		},
	}
	mockLogger := new(logger.MockLogger)

	t.Run("SendMessage", func(t *testing.T) {
		mockProducer := new(MockSyncProducer)
		service := KafkaService{
			Producer: mockProducer,
			Config:   cfg,
			Log:      mockLogger,
		}

		message := &model.MediaProcessingMessage{VideoID: "test-video"}
		expectedPartition := int32(0)
		expectedOffset := int64(1)

		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockProducer.On("SendMessage", mock.Anything).Return(expectedPartition, expectedOffset, nil)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()

		// act
		err := service.SendMessage(context.Background(), message)

		// assert
		assert.NoError(t, err)
		mockProducer.AssertExpectations(t)
	})

	t.Run("ReceiveMessage", func(t *testing.T) {
		// arrange
		mockConsumer := new(MockConsumer)
		mockPartitionConsumer := new(MockPartitionConsumer)

		service := KafkaService{
			Config:   cfg,
			Log:      mockLogger,
			Consumer: mockConsumer,
		}

		message := model.MediaProcessingMessage{VideoID: "test-video"}
		messageBytes, err := json.Marshal(message)
		assert.NoError(t, err)

		messageChan := make(chan *sarama.ConsumerMessage, 1)
		messageChan <- &sarama.ConsumerMessage{Value: messageBytes}

		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockConsumer.On("ConsumePartition", cfg.Messaging.Kafka.Topic, int32(0), sarama.OffsetOldest).Return(mockPartitionConsumer, nil)
		mockPartitionConsumer.On("Messages").Return((<-chan *sarama.ConsumerMessage)(messageChan))
		mockPartitionConsumer.On("Close").Return(nil)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()

		// act
		messages, err := service.ReceiveMessage(context.Background())

		// assert
		assert.NoError(t, err)
		assert.Len(t, messages, 1)
		assert.Equal(t, message.VideoID, messages[0].VideoID)

		mockConsumer.AssertExpectations(t)
		mockPartitionConsumer.AssertExpectations(t)
	})

	t.Run("DeleteMessage", func(t *testing.T) {
		// arrange
		service := KafkaService{
			Config: cfg,
			Log:    mockLogger,
		}

		mockLogger.On("With", mock.Anything).Return(mockLogger)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return()

		// act
		err := service.DeleteMessage(context.Background(), "test-id")

		assert.NoError(t, err)
	})

}
