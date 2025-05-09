//go:build !integration

package worker

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestDownloadWorkerPool_Start_Success(t *testing.T) {
	mockProcessor := new(MockProcessor)
	mockLogger := new(logger.MockLogger)
	mockConsumer := new(MockConsumer)
	mockFactory := new(MockWorkerFactory)
	mockWorker := new(MockTaskWorker)

	ctx, cancel := context.WithCancel(context.Background())
	requestChan := make(chan *model.MediaRequest)

	numWorkers := 2

	mockConsumer.On("GetRequestsChannel", ctx).Return((<-chan *model.MediaRequest)(requestChan), nil)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	// Expectativas para la fábrica
	for i := 0; i < numWorkers; i++ {
		mockFactory.On("NewWorker", i, mockProcessor, mockLogger).Return(mockWorker)
		mockWorker.On("Run", ctx, mock.AnythingOfType("*sync.WaitGroup"), (<-chan *model.MediaRequest)(requestChan)).Return()
	}

	workerPool := NewDownloadWorkerPool(
		numWorkers,
		mockConsumer,
		mockProcessor,
		mockLogger,
		mockFactory,
	)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := workerPool.Start(ctx)

	assert.NoError(t, err)
	mockConsumer.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	mockFactory.AssertExpectations(t)
	mockWorker.AssertExpectations(t)
}

func TestDownloadWorkerPool_Start_GetRequestsChannelError(t *testing.T) {
	mockProcessor := new(MockProcessor)
	mockLogger := new(logger.MockLogger)
	mockConsumer := new(MockConsumer)
	mockFactory := new(MockWorkerFactory)

	ctx := context.Background()
	expectedError := errors.New("error al obtener canal de solicitudes")

	mockConsumer.On("GetRequestsChannel", ctx).Return((<-chan *model.MediaRequest)(nil), expectedError)

	workerPool := NewDownloadWorkerPool(
		2,
		mockConsumer,
		mockProcessor,
		mockLogger,
		mockFactory,
	)

	err := workerPool.Start(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al obtener el canal de solicitudes")
	mockConsumer.AssertExpectations(t)
}

func TestNewDownloadWorkerPool_Initialization(t *testing.T) {
	mockProcessor := new(MockProcessor)
	mockLogger := new(logger.MockLogger)
	mockConsumer := new(MockConsumer)
	mockFactory := new(MockWorkerFactory)

	numWorkers := 3

	workerPool := NewDownloadWorkerPool(
		numWorkers,
		mockConsumer,
		mockProcessor,
		mockLogger,
		mockFactory,
	)

	assert.NotNil(t, workerPool)
	assert.Equal(t, numWorkers, workerPool.workerCount)
	assert.Equal(t, mockConsumer, workerPool.consumer)
	assert.Equal(t, mockProcessor, workerPool.processor)
	assert.Equal(t, mockLogger, workerPool.logger)
	assert.Equal(t, mockFactory, workerPool.workerFactory)
}
