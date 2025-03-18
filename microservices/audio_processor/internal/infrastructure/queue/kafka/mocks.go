package kafka

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/mock"
)

type (
	MockSyncProducer struct {
		mock.Mock
	}

	MockConsumer struct {
		mock.Mock
	}

	MockPartitionConsumer struct {
		mock.Mock
	}
)

func (m *MockSyncProducer) TxnStatus() sarama.ProducerTxnStatusFlag {
	return m.Called().Get(0).(sarama.ProducerTxnStatusFlag)
}

func (m *MockSyncProducer) IsTransactional() bool {
	return m.Called().Bool(0)

}

func (m *MockSyncProducer) BeginTxn() error {
	return m.Called().Error(0)

}

func (m *MockSyncProducer) CommitTxn() error {
	return m.Called().Error(0)

}

func (m *MockSyncProducer) AbortTxn() error {
	return m.Called().Error(0)
}

func (m *MockSyncProducer) AddOffsetsToTxn(offsets map[string][]*sarama.PartitionOffsetMetadata, groupId string) error {
	args := m.Called(offsets, groupId)
	return args.Error(0)
}

func (m *MockSyncProducer) AddMessageToTxn(msg *sarama.ConsumerMessage, groupId string, metadata *string) error {
	args := m.Called(msg, groupId, metadata)
	return args.Error(0)
}

func (m *MockSyncProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	args := m.Called(msg)
	return args.Get(0).(int32), args.Get(1).(int64), args.Error(2)
}

func (m *MockSyncProducer) SendMessages(msgs []*sarama.ProducerMessage) error {
	args := m.Called(msgs)
	return args.Error(0)
}

func (m *MockSyncProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockConsumer) Topics() ([]string, error) {
	return m.Called().Get(0).([]string), nil
}

func (m *MockConsumer) Partitions(topic string) ([]int32, error) {
	args := m.Called(topic)
	return args.Get(0).([]int32), args.Error(1)
}

func (m *MockConsumer) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	args := m.Called(topic, partition, offset)
	return args.Get(0).(sarama.PartitionConsumer), args.Error(1)
}

func (m *MockConsumer) HighWaterMarks() map[string]map[int32]int64 {
	return m.Called().Get(0).(map[string]map[int32]int64)
}

func (m *MockConsumer) Close() error {
	return m.Called().Error(0)
}

func (m *MockConsumer) Pause(topicPartitions map[string][]int32) {
	m.Called(topicPartitions)
}

func (m *MockConsumer) Resume(topicPartitions map[string][]int32) {
	m.Called(topicPartitions)
}

func (m *MockConsumer) PauseAll() {
}

func (m *MockConsumer) ResumeAll() {
}

func (m *MockPartitionConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockPartitionConsumer) AsyncClose() {}

func (m *MockPartitionConsumer) Messages() <-chan *sarama.ConsumerMessage {
	args := m.Called()
	return args.Get(0).(<-chan *sarama.ConsumerMessage)
}

func (m *MockPartitionConsumer) Errors() <-chan *sarama.ConsumerError {
	return m.Called().Get(0).(chan *sarama.ConsumerError)
}

func (m *MockPartitionConsumer) HighWaterMarkOffset() int64 {
	return m.Called().Get(0).(int64)
}

func (m *MockPartitionConsumer) Pause() {}

func (m *MockPartitionConsumer) Resume() {}

func (m *MockPartitionConsumer) IsPaused() bool {
	return m.Called().Get(0).(bool)
}
