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

func TestWorkerPool_Start_Success(t *testing.T) {
	mockProcessor := new(MockProcessor)
	mockLogger := new(logger.MockLogger)
	mockConsumer := new(MockConsumer)

	ctx, cancel := context.WithCancel(context.Background())
	requestChan := make(chan *model.MediaRequest)

	numWorkers := 2

	mockConsumer.On("GetRequestsChannel", ctx).Return((<-chan *model.MediaRequest)(requestChan), nil)
	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	workerPool := NewWorkerPool(numWorkers, mockConsumer, mockProcessor, mockLogger)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := workerPool.Start(ctx)

	assert.NoError(t, err)
	mockConsumer.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
	assert.Len(t, workerPool.workers, numWorkers)
}

func TestWorkerPool_Start_GetRequestsChannelError(t *testing.T) {
	mockProcessor := new(MockProcessor)
	mockLogger := new(logger.MockLogger)
	mockConsumer := new(MockConsumer)

	ctx := context.Background()
	expectedError := errors.New("error al obtener canal de solicitudes")

	mockConsumer.On("GetRequestsChannel", ctx).Return((<-chan *model.MediaRequest)(nil), expectedError)

	workerPool := NewWorkerPool(2, mockConsumer, mockProcessor, mockLogger)

	err := workerPool.Start(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error al obtener el canal de solicitudes")
	mockConsumer.AssertExpectations(t)
}

func TestNewWorkerPool_Initialization(t *testing.T) {
	mockProcessor := new(MockProcessor)
	mockLogger := new(logger.MockLogger)
	mockConsumer := new(MockConsumer)

	numWorkers := 3

	workerPool := NewWorkerPool(numWorkers, mockConsumer, mockProcessor, mockLogger)

	assert.NotNil(t, workerPool)
	assert.Equal(t, numWorkers, workerPool.numWorkers)
	assert.Equal(t, mockConsumer, workerPool.consumer)
	assert.Equal(t, mockProcessor, workerPool.processor)
	assert.Equal(t, mockLogger, workerPool.logger)
	assert.Len(t, workerPool.workers, numWorkers)
}
