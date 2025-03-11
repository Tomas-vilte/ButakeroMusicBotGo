package usecase

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/stretchr/testify/mock"
)

type MockMediaRepository struct {
	mock.Mock
}

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
