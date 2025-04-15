package service

import (
	"bytes"
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/stretchr/testify/mock"
	"io"
)

type (
	MockVideoProvider struct {
		mock.Mock
	}

	MockMessageQueue struct {
		mock.Mock
	}

	MockMediaRepository struct {
		mock.Mock
	}

	MockStorage struct {
		mock.Mock
	}

	MockMediaService struct {
		mock.Mock
	}

	MockAudioStorageService struct {
		mock.Mock
	}

	MockTopicPublisherService struct {
		mock.Mock
	}

	MockAudioDownloadService struct {
		mock.Mock
	}
)

func (m *MockVideoProvider) GetVideoDetails(ctx context.Context, videoID string) (*model.MediaDetails, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).(*model.MediaDetails), args.Error(1)
}

func (m *MockVideoProvider) SearchVideoID(ctx context.Context, input string) (string, error) {
	args := m.Called(ctx, input)
	return args.String(0), args.Error(1)
}

func (m *MockMessageQueue) Publish(ctx context.Context, message *model.MediaProcessingMessage) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockMessageQueue) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMediaRepository) SaveMedia(ctx context.Context, media *model.Media) error {
	args := m.Called(ctx, media)
	return args.Error(0)
}

func (m *MockMediaRepository) GetMediaByID(ctx context.Context, videoID string) (*model.Media, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).(*model.Media), args.Error(1)
}

func (m *MockMediaRepository) GetMediaByTitle(ctx context.Context, title string) ([]*model.Media, error) {
	args := m.Called(ctx, title)
	return args.Get(0).([]*model.Media), args.Error(1)
}

func (m *MockMediaRepository) DeleteMedia(ctx context.Context, videoID string) error {
	args := m.Called(ctx, videoID)
	return args.Error(0)
}

func (m *MockMediaRepository) UpdateMedia(ctx context.Context, videoID string, media *model.Media) error {
	args := m.Called(ctx, videoID, media)
	return args.Error(0)
}

func (m *MockStorage) UploadFile(ctx context.Context, key string, body io.Reader) error {
	args := m.Called(ctx, key, body)
	return args.Error(0)
}

func (m *MockStorage) GetFileMetadata(ctx context.Context, key string) (*model.FileData, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*model.FileData), args.Error(1)
}

func (m *MockStorage) GetFileContent(ctx context.Context, path string, key string) (io.ReadCloser, error) {
	args := m.Called(ctx, path, key)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockMediaService) CreateMedia(ctx context.Context, media *model.Media) error {
	args := m.Called(ctx, media)
	return args.Error(0)
}
func (m *MockMediaService) GetMediaByID(ctx context.Context, videoID string) (*model.Media, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).(*model.Media), args.Error(1)
}

func (m *MockMediaService) UpdateMedia(ctx context.Context, videoID string, status *model.Media) error {
	args := m.Called(ctx, videoID, status)
	return args.Error(0)
}

func (m *MockMediaService) GetMediaByTitle(ctx context.Context, title string) ([]*model.Media, error) {
	args := m.Called(ctx, title)
	return args.Get(0).([]*model.Media), args.Error(1)
}

func (m *MockMediaService) DeleteMedia(ctx context.Context, videoID string) error {
	args := m.Called(ctx, videoID)
	return args.Error(0)
}

func (m *MockAudioStorageService) StoreAudio(ctx context.Context, buffer *bytes.Buffer, songName string) (*model.FileData, error) {
	args := m.Called(ctx, buffer, songName)
	return args.Get(0).(*model.FileData), args.Error(1)
}

func (m *MockTopicPublisherService) PublishMediaProcessed(ctx context.Context, message *model.MediaProcessingMessage) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockAudioDownloadService) DownloadAndEncode(ctx context.Context, url string) (*bytes.Buffer, error) {
	args := m.Called(ctx, url)
	return args.Get(0).(*bytes.Buffer), args.Error(1)
}
