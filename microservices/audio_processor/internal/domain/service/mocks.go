package service

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/stretchr/testify/mock"
)

type (
	MockVideoProvider struct {
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
