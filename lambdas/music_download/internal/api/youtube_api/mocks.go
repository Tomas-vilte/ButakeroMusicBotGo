package youtube_api

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/types"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"google.golang.org/api/youtube/v3"
)

type MockSongLooker struct {
	mock.Mock
}

func (m *MockSongLooker) LookupSongs(ctx context.Context, input string) ([]*types.Song, error) {
	args := m.Called(ctx, input)
	return args.Get(0).([]*types.Song), args.Error(1)
}

func (m *MockSongLooker) SearchYouTubeVideoID(ctx context.Context, searchTerm string) (string, error) {
	args := m.Called(ctx, searchTerm)
	return args.String(0), args.Error(1)
}

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

type MockYouTubeService struct {
	mock.Mock
}

func (m *MockYouTubeService) GetVideoDetails(ctx context.Context, videoID string) (*youtube.Video, error) {
	args := m.Called(ctx, videoID)
	return args.Get(0).(*youtube.Video), args.Error(1)
}

func (m *MockYouTubeService) SearchVideoID(ctx context.Context, searchTerm string) (string, error) {
	args := m.Called(ctx, searchTerm)
	return args.String(0), args.Error(1)
}
