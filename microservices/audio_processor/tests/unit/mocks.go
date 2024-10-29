package unit

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/port"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
	"io"
)

type (
	MockOperationRepository struct {
		mock.Mock
	}

	MockMetadataRepository struct {
		mock.Mock
	}

	MockDownloader struct {
		mock.Mock
	}

	MockStorage struct {
		mock.Mock
	}

	MockLogger struct {
		mock.Mock
	}

	MockYouTubeService struct {
		mock.Mock
	}

	MockAudioProcessingService struct {
		mock.Mock
	}

	MockInitiateDownloadUC struct {
		mock.Mock
	}

	MockGetOperationStatusUC struct {
		mock.Mock
	}

	MockStorageS3API struct {
		mock.Mock
	}

	MockDynamoDBAPI struct {
		mock.Mock
	}

	MockSQSClient struct {
		mock.Mock
	}

	MockSyncProducer struct {
		mock.Mock
	}

	MockConsumer struct {
		mock.Mock
	}

	MockPartitionConsumer struct {
		mock.Mock
	}

	MockMessagingQueue struct {
		mock.Mock
	}
)

func (m *MockMessagingQueue) SendMessage(ctx context.Context, message port.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessagingQueue) ReceiveMessage(ctx context.Context) ([]port.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).([]port.Message), args.Error(1)
}

func (m *MockMessagingQueue) DeleteMessage(ctx context.Context, receiptHandle string) error {
	args := m.Called(ctx, receiptHandle)
	return args.Error(0)
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
	return
}

func (m *MockConsumer) Resume(topicPartitions map[string][]int32) {
	return
}

func (m *MockConsumer) PauseAll() {
	return
}

func (m *MockConsumer) ResumeAll() {
	return
}

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

func (m *MockSQSClient) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*sqs.SendMessageOutput), args.Error(1)
}

func (m *MockSQSClient) ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*sqs.ReceiveMessageOutput), args.Error(1)
}

func (m *MockSQSClient) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

func (m *MockStorageS3API) HeadObject(ctx context.Context, params *s3.HeadObjectInput, optFns ...func(*s3.Options)) (*s3.HeadObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.HeadObjectOutput), args.Error(1)
}

func (m *MockStorageS3API) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*s3.PutObjectOutput), args.Error(1)
}

func (m *MockOperationRepository) SaveOperationsResult(ctx context.Context, result *model.OperationResult) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}

func (m *MockOperationRepository) GetOperationResult(ctx context.Context, id, songID string) (*model.OperationResult, error) {
	args := m.Called(ctx, id, songID)
	return args.Get(0).(*model.OperationResult), args.Error(1)
}

func (m *MockOperationRepository) DeleteOperationResult(ctx context.Context, id, songID string) error {
	args := m.Called(ctx, id, songID)
	return args.Error(0)
}

func (m *MockOperationRepository) UpdateOperationStatus(ctx context.Context, operationID, songID, status string) error {
	args := m.Called(ctx, operationID, songID, status)
	return args.Error(0)
}

func (m *MockOperationRepository) UpdateOperationResult(ctx context.Context, operationID string, operationResult *model.OperationResult) error {
	args := m.Called(ctx, operationID, operationResult)
	return args.Error(0)
}

func (m *MockMetadataRepository) SaveMetadata(ctx context.Context, metadata *model.Metadata) error {
	args := m.Called(ctx, metadata)
	return args.Error(0)
}

func (m *MockMetadataRepository) GetMetadata(ctx context.Context, id string) (*model.Metadata, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Metadata), args.Error(1)
}

func (m *MockMetadataRepository) DeleteMetadata(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockDownloader) DownloadAudio(ctx context.Context, url string) (io.Reader, error) {
	args := m.Called(ctx, url)
	return args.Get(0).(io.Reader), args.Error(1)
}

func (m *MockStorage) UploadFile(ctx context.Context, key string, body io.Reader) error {
	args := m.Called(ctx, key, body)
	return args.Error(0)
}

func (m *MockStorage) GetFileMetadata(ctx context.Context, key string) (*model.FileData, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*model.FileData), args.Error(1)
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

func (m *MockYouTubeService) SearchVideoID(ctx context.Context, song string) (string, error) {
	args := m.Called(ctx, song)
	return args.String(0), args.Error(1)
}

func (m *MockYouTubeService) GetVideoDetails(ctx context.Context, videoID string) (*api.VideoDetails, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).(*api.VideoDetails), args.Error(1)
}

func (m *MockAudioProcessingService) StartOperation(ctx context.Context, videoID string) (string, string, error) {
	args := m.Called(ctx, videoID)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockAudioProcessingService) ProcessAudio(ctx context.Context, operationID string, videoDetails *api.VideoDetails) error {
	args := m.Called(ctx, operationID, videoDetails)
	return args.Error(0)
}

func (m *MockInitiateDownloadUC) Execute(ctx context.Context, song string) (string, string, error) {
	args := m.Called(ctx, song)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockGetOperationStatusUC) Execute(ctx context.Context, operationID, songID string) (*model.OperationResult, error) {
	args := m.Called(ctx, operationID, songID)
	return args.Get(0).(*model.OperationResult), args.Error(1)
}

func (m *MockDynamoDBAPI) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *MockDynamoDBAPI) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *MockDynamoDBAPI) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.DeleteItemOutput), args.Error(1)
}

func (m *MockDynamoDBAPI) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	return args.Get(0).(*dynamodb.UpdateItemOutput), args.Error(1)
}
