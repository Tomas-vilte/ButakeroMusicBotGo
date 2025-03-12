package usecase

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/stretchr/testify/mock"
)

type (
	MockMediaRepository struct {
		mock.Mock
	}

	MockCoreService struct {
		mock.Mock
	}

	MockVideoService struct {
		mock.Mock
	}

	MockOperationService struct {
		mock.Mock
	}
)

func (m *MockMediaRepository) SaveMedia(ctx context.Context, media *model.Media) error {
	args := m.Called(ctx, media)
	return args.Error(0)
}

func (m *MockMediaRepository) GetMedia(ctx context.Context, id string, videoID string) (*model.Media, error) {
	args := m.Called(ctx, id, videoID)
	return args.Get(0).(*model.Media), args.Error(1)
}

func (m *MockMediaRepository) DeleteMedia(ctx context.Context, id string, videoID string) error {
	args := m.Called(ctx, id, videoID)
	return args.Error(0)
}

func (m *MockMediaRepository) UpdateMedia(ctx context.Context, id string, videoID string, media *model.Media) error {
	args := m.Called(ctx, id, videoID, media)
	return args.Error(0)
}

func (m *MockCoreService) ProcessMedia(ctx context.Context, operationID string, media *model.MediaDetails) error {
	args := m.Called(ctx, operationID, media)
	return args.Error(0)
}

func (m *MockVideoService) GetMediaDetails(ctx context.Context, song string, providerType string) (*model.MediaDetails, error) {
	args := m.Called(ctx, song, providerType)
	return args.Get(0).(*model.MediaDetails), args.Error(1)
}

func (m *MockOperationService) StartOperation(ctx context.Context, mediaID string) (*model.OperationInitResult, error) {
	args := m.Called(ctx, mediaID)
	return args.Get(0).(*model.OperationInitResult), args.Error(1)
}
