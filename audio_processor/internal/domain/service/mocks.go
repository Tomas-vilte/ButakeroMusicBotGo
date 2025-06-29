package service

import (
	"bytes"
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
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

	MockAudioStorageService struct {
		mock.Mock
	}

	MockAudioDownloadService struct {
		mock.Mock
	}

	MockDownloader struct {
		mock.Mock
	}

	MockAudioEncoder struct {
		mock.Mock
	}

	MockEncodeSession struct {
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

func (m *MockAudioStorageService) StoreAudio(ctx context.Context, buffer *bytes.Buffer, songName string) (*model.FileData, error) {
	args := m.Called(ctx, buffer, songName)
	return args.Get(0).(*model.FileData), args.Error(1)
}

func (m *MockAudioDownloadService) DownloadAndEncode(ctx context.Context, url string) (*bytes.Buffer, error) {
	args := m.Called(ctx, url)
	return args.Get(0).(*bytes.Buffer), args.Error(1)
}

func (m *MockDownloader) DownloadAudio(ctx context.Context, url string) (io.Reader, error) {
	args := m.Called(ctx, url)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(io.Reader), args.Error(1)
}

func (m *MockAudioEncoder) Encode(ctx context.Context, reader io.Reader, options *model.EncodeOptions) (ports.EncodeSession, error) {
	args := m.Called(ctx, reader, options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(ports.EncodeSession), args.Error(1)
}

func (m *MockEncodeSession) ReadFrame() ([]byte, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEncodeSession) Read(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *MockEncodeSession) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockEncodeSession) FFMPEGMessages() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockEncodeSession) Cleanup() {
	m.Called()
}
