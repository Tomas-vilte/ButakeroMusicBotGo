package handler

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/types"
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

type MockSongLooker struct {
	mock.Mock
}

func (m *MockSongLooker) SearchYouTubeVideoID(ctx context.Context, query string) (string, error) {
	args := m.Called(ctx, query)
	return args.String(0), args.Error(1)
}

func (m *MockSongLooker) LookupSongs(ctx context.Context, videoID string) ([]*types.Song, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).([]*types.Song), args.Error(1)
}

type MockSQSClient struct {
	mock.Mock
}

func (m *MockSQSClient) SendMessage(ctx context.Context, messageBody string) error {
	args := m.Called(ctx, messageBody)
	return args.Error(0)
}
