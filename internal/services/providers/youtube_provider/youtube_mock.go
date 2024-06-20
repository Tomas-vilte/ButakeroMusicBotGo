package youtube_provider

import (
	"context"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
	"google.golang.org/api/youtube/v3"
)

type MockYouTubeClient struct {
	mock.Mock
}

func (m *MockYouTubeClient) VideosListCall(ctx context.Context, part []string) VideosListCallWrapper {
	args := m.Called(ctx, part)
	return args.Get(0).(VideosListCallWrapper)
}

func (m *MockYouTubeClient) SearchListCall(ctx context.Context, part []string) SearchListCallWrapper {
	args := m.Called(ctx, part)
	return args.Get(0).(SearchListCallWrapper)
}

type MockLogger struct {
	mock.Mock
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

type SearchListCallWrapperMock struct {
	mock.Mock
}

type VideosListCallWrapperMock struct {
	mock.Mock
}

func (m *VideosListCallWrapperMock) Id(id string) VideosListCallWrapper {
	args := m.Called(id)
	return args.Get(0).(VideosListCallWrapper)
}

func (m *VideosListCallWrapperMock) Do() (*youtube.VideoListResponse, error) {
	args := m.Called()
	return args.Get(0).(*youtube.VideoListResponse), args.Error(1)
}

func (m *SearchListCallWrapperMock) Q(q string) SearchListCallWrapper {
	args := m.Called(q)
	return args.Get(0).(SearchListCallWrapper)
}

func (m *SearchListCallWrapperMock) MaxResults(maxResults int64) SearchListCallWrapper {
	args := m.Called(maxResults)
	return args.Get(0).(SearchListCallWrapper)
}

func (m *SearchListCallWrapperMock) Type(typ string) SearchListCallWrapper {
	args := m.Called(typ)
	return args.Get(0).(SearchListCallWrapper)
}

func (m *SearchListCallWrapperMock) Do() (*youtube.SearchListResponse, error) {
	args := m.Called()
	return args.Get(0).(*youtube.SearchListResponse), args.Error(1)
}
