package worker

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/stretchr/testify/mock"
)

type MockProcessor struct {
	mock.Mock
}

func (m *MockProcessor) ProcessRequest(ctx context.Context, req *model.MediaRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

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
