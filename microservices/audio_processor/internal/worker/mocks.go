package worker

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/mock"
	"sync"
)

type MockProcessor struct {
	mock.Mock
}

func (m *MockProcessor) ProcessDownloadTask(ctx context.Context, req *model.MediaRequest) error {
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

type MockTaskWorker struct {
	mock.Mock
}

func (m *MockTaskWorker) Run(ctx context.Context, wg *sync.WaitGroup, taskChan <-chan *model.MediaRequest) {
	m.Called(ctx, wg, taskChan)
	wg.Done()
}

type MockWorkerFactory struct {
	mock.Mock
}

func (m *MockWorkerFactory) NewWorker(id int, processor AudioTaskProcessor, logger logger.Logger) TaskWorker {
	args := m.Called(id, processor, logger)
	return args.Get(0).(TaskWorker)
}
