package service

import (
	"bytes"
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/stretchr/testify/mock"
	"io"
)

type (
	MockAudioDownloadService struct {
		mock.Mock
	}

	MockAudioStorageService struct {
		mock.Mock
	}

	MockOperationsManager struct {
		mock.Mock
	}

	MockMessagingManager struct {
		mock.Mock
	}

	MockErrorManagement struct {
		mock.Mock
	}

	MockMetadataRepository struct {
		mock.Mock
	}

	MockStorage struct {
		mock.Mock
	}

	MockVideoProvider struct {
		mock.Mock
	}

	MockOperationRepository struct {
		mock.Mock
	}

	MockMessagingQueue struct {
		mock.Mock
	}
)

func (m *MockAudioDownloadService) DownloadAndEncode(ctx context.Context, url string) (*bytes.Buffer, error) {
	args := m.Called(ctx, url)
	return args.Get(0).(*bytes.Buffer), args.Error(1)
}

func (m *MockAudioStorageService) StoreAudio(ctx context.Context, buffer *bytes.Buffer, metadata *model.Metadata) (*model.FileData, error) {
	args := m.Called(ctx, buffer, metadata)
	return args.Get(0).(*model.FileData), args.Error(1)
}

func (m *MockOperationsManager) HandleOperationSuccess(ctx context.Context, operationID string, metadata *model.Metadata, fileData *model.FileData) error {
	args := m.Called(ctx, operationID, metadata, fileData)
	return args.Error(0)
}

func (m *MockMessagingManager) SendProcessingMessage(ctx context.Context, operationID string, status string, metadata *model.Metadata, attempts int) error {
	args := m.Called(ctx, operationID, status, metadata, attempts)
	return args.Error(0)
}

func (m *MockErrorManagement) HandleProcessingError(ctx context.Context, operationID string, metadata *model.Metadata, stage string, attempts int, err error) error {
	args := m.Called(ctx, operationID, metadata, stage, attempts, err)
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

func (m *MockStorage) UploadFile(ctx context.Context, key string, body io.Reader) error {
	args := m.Called(ctx, key, body)
	return args.Error(0)
}

func (m *MockStorage) GetFileMetadata(ctx context.Context, keyName string) (*model.FileData, error) {
	args := m.Called(ctx, keyName)
	return args.Get(0).(*model.FileData), args.Error(1)
}

func (m *MockStorage) GetFileContent(ctx context.Context, path string, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, path, key)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockVideoProvider) GetVideoDetails(ctx context.Context, videoID string) (*model.MediaDetails, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).(*model.MediaDetails), args.Error(1)
}

func (m *MockVideoProvider) SearchVideoID(ctx context.Context, input string) (string, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(string), args.Error(1)
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

func (m *MockOperationRepository) UpdateOperationStatus(ctx context.Context, operationID string, songID string, status string) error {
	args := m.Called(ctx, operationID, songID, status)
	return args.Error(0)
}

func (m *MockOperationRepository) UpdateOperationResult(ctx context.Context, operationID string, operationResult *model.OperationResult) error {
	args := m.Called(ctx, operationID, operationResult)
	return args.Error(0)
}

func (m *MockMessagingQueue) SendMessage(ctx context.Context, message model.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessagingQueue) ReceiveMessage(ctx context.Context) ([]model.Message, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Message), args.Error(1)
}
func (m *MockMessagingQueue) DeleteMessage(ctx context.Context, receiptHandle string) error {
	args := m.Called(ctx, receiptHandle)
	return args.Error(0)
}
