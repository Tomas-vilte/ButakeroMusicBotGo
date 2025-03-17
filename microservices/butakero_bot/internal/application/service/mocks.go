package service

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/stretchr/testify/mock"
)

type MockSongDownloader struct {
	mock.Mock
}

func (m *MockSongDownloader) DownloadSong(ctx context.Context, songName, providerType string) (*entity.DownloadResponse, error) {
	args := m.Called(ctx, songName, providerType)
	return args.Get(0).(*entity.DownloadResponse), args.Error(1)
}
