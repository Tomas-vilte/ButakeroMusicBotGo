package handler

import (
	"context"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
	"io"
)

// MockDownloader es un mock para la interfaz Downloader
type MockDownloader struct {
	mock.Mock
}

func (m *MockDownloader) DownloadSong(songURL string, key string) error {
	args := m.Called(songURL, key)
	return args.Error(0)
}

// MockUploader es un mock para la interfaz Uploader
type MockUploader struct {
	mock.Mock
}

type MockLogger struct {
	mock.Mock
}

func (m *MockUploader) UploadToS3(ctx context.Context, reader io.Reader, key string) error {
	args := m.Called(ctx, reader, key)
	return args.Error(0)
}

func (m *MockLogger) Info(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) With(fields ...zapcore.Field) {
	m.Called(fields)
}
