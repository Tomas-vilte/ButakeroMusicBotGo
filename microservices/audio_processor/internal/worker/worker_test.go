//go:build !integration

package worker

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/mock"
	"sync"
	"testing"
)

func TestDownloadTaskWorker_Run_ProcessesTask(t *testing.T) {
	mockProcessor := new(MockProcessor)
	mockLogger := new(logger.MockLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskChan := make(chan *model.MediaRequest, 1)
	var wg sync.WaitGroup
	wg.Add(1)

	worker := NewDownloadTaskWorker(1, mockProcessor, mockLogger)

	task := &model.MediaRequest{
		RequestID: "test-request-id",
		Song:      "https://test-url.com",
	}

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockProcessor.On("ProcessDownloadTask", ctx, task).Return(nil)

	go worker.Run(ctx, &wg, taskChan)
	taskChan <- task

	close(taskChan)

	wg.Wait()

	mockProcessor.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestDownloadTaskWorker_Run_HandlesProcessingError(t *testing.T) {
	mockProcessor := new(MockProcessor)
	mockLogger := new(logger.MockLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	taskChan := make(chan *model.MediaRequest, 1)
	var wg sync.WaitGroup
	wg.Add(1)

	worker := NewDownloadTaskWorker(1, mockProcessor, mockLogger)

	task := &model.MediaRequest{
		RequestID: "test-request-id",
		Song:      "https://test-url.com",
	}

	expectedError := errors.New("error procesando tarea")

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockProcessor.On("ProcessDownloadTask", ctx, task).Return(expectedError)

	go worker.Run(ctx, &wg, taskChan)
	taskChan <- task

	close(taskChan)

	wg.Wait()

	mockProcessor.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestDownloadTaskWorker_Run_HandlesContextCancellation(t *testing.T) {
	mockProcessor := new(MockProcessor)
	mockLogger := new(logger.MockLogger)

	ctx, cancel := context.WithCancel(context.Background())

	taskChan := make(chan *model.MediaRequest)
	var wg sync.WaitGroup
	wg.Add(1)

	worker := NewDownloadTaskWorker(1, mockProcessor, mockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	go worker.Run(ctx, &wg, taskChan)

	cancel()

	wg.Wait()

	mockLogger.AssertExpectations(t)
}
