package unit

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
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
)

func (m *MockOperationRepository) SaveOperationsResult(ctx context.Context, result model.OperationResult) error {
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

func (m *MockMetadataRepository) SaveMetadata(ctx context.Context, metadata model.Metadata) error {
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

func (m *MockAudioProcessingService) StartOperation(ctx context.Context, videoID string) (string, error) {
	args := m.Called(ctx, videoID)
	return args.String(0), args.Error(1)
}

func (m *MockAudioProcessingService) ProcessAudio(ctx context.Context, operationID string, videoDetails api.VideoDetails) error {
	args := m.Called(ctx, operationID, videoDetails)
	return args.Error(0)
}
