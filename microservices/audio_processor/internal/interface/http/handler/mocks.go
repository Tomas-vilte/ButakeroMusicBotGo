package handler

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/stretchr/testify/mock"
)

type MockInitiateDownloadUC struct {
	mock.Mock
}

func (m *MockInitiateDownloadUC) Execute(ctx context.Context, song string, providerType string) (*model.OperationInitResult, error) {
	args := m.Called(ctx, song, providerType)
	return args.Get(0).(*model.OperationInitResult), args.Error(1)
}
