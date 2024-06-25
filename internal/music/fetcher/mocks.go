package fetcher

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/api/youtube/v3"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) With(fields ...zap.Field) {
	m.Called(fields)
}

type MockAudioCaching struct {
	mock.Mock
}

func (m *MockAudioCaching) Get(url string) ([]byte, bool) {
	args := m.Called(url)
	data, _ := args.Get(0).([]byte)
	found, _ := args.Get(1).(bool)
	return data, found
}

func (m *MockAudioCaching) Set(url string, data []byte) {
	m.Called(url, data)
}

func (m *MockAudioCaching) Size() int {
	args := m.Called()
	size, _ := args.Get(0).(int)
	return size
}

// MockCacheManager es un mock de la interfaz Manager
type MockCacheManager struct {
	mock.Mock
}

func (m *MockCacheManager) Get(key string) []*voice.Song {
	args := m.Called(key)
	results, _ := args.Get(0).([]*voice.Song)
	return results
}

func (m *MockCacheManager) Set(key string, results []*voice.Song) {
	m.Called(key, results)
}

func (m *MockCacheManager) DeleteExpiredEntries() {
	m.Called()
}

func (m *MockCacheManager) Size() int {
	args := m.Called()
	size, _ := args.Get(0).(int)
	return size
}

type MockYouTubeService struct {
	mock.Mock
}

func (m *MockYouTubeService) SearchVideoID(ctx context.Context, searchTerm string) (string, error) {
	args := m.Called(ctx, searchTerm)
	return args.String(0), args.Error(1)
}

func (m *MockYouTubeService) GetVideoDetails(ctx context.Context, videoID string) (*youtube.Video, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).(*youtube.Video), args.Error(1)
}

type MockCommandExecutor struct {
	mock.Mock
}

func (m *MockCommandExecutor) ExecuteCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	argsMock := m.Called(ctx, name, args)
	data, _ := argsMock.Get(0).([]byte)
	err := argsMock.Error(1)
	return data, err
}
