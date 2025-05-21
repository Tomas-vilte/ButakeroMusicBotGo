package processor

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/stretchr/testify/mock"
)

type MockConsumer struct {
	mock.Mock
}

func (m *MockConsumer) GetRequestsChannel(ctx context.Context) (<-chan *model.MediaRequest, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan *model.MediaRequest), args.Error(1)
}

func (m *MockConsumer) Close() error {
	m.Called()
	return nil
}

type MockMediaRepository struct {
	mock.Mock
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

type MockVideoService struct {
	mock.Mock
}

func (m *MockVideoService) GetMediaDetails(ctx context.Context, input string, providerType string) (*model.MediaDetails, error) {
	args := m.Called(ctx, input, providerType)
	return args.Get(0).(*model.MediaDetails), args.Error(1)
}

type MockCoreService struct {
	mock.Mock
}

func (m *MockCoreService) ProcessMedia(ctx context.Context, media *model.Media, userID, requestID string) error {
	args := m.Called(ctx, media, userID, requestID)
	return args.Error(0)
}

type MockDownloadSongWorkerPool struct {
	mock.Mock
}

func (m *MockDownloadSongWorkerPool) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
