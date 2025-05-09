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

func TestWorker_Start_ProcessesRequest(t *testing.T) {
	mockProcessor := new(MockProcessor)
	mockLogger := new(logger.MockLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	requestChan := make(chan *model.MediaRequest, 1)
	var wg sync.WaitGroup
	wg.Add(1)

	worker := NewWorker(1, mockProcessor, mockLogger)

	request := &model.MediaRequest{
		RequestID: "test-request-id",
		Song:      "https://test-url.com",
	}

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockProcessor.On("ProcessRequest", ctx, request).Return(nil)

	go worker.Start(ctx, &wg, requestChan)
	requestChan <- request

	close(requestChan)

	wg.Wait()

	mockProcessor.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestWorker_Start_HandlesProcessingError(t *testing.T) {
	mockProcessor := new(MockProcessor)
	mockLogger := new(logger.MockLogger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	requestChan := make(chan *model.MediaRequest, 1)
	var wg sync.WaitGroup
	wg.Add(1)

	worker := NewWorker(1, mockProcessor, mockLogger)

	request := &model.MediaRequest{
		RequestID: "test-request-id",
		Song:      "https://test-url.com",
	}

	expectedError := errors.New("error procesando solicitud")

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()
	mockLogger.On("Error", mock.Anything, mock.Anything).Return()
	mockProcessor.On("ProcessRequest", ctx, request).Return(expectedError)

	go worker.Start(ctx, &wg, requestChan)
	requestChan <- request

	close(requestChan)

	wg.Wait()

	mockProcessor.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}

func TestWorker_Start_HandlesContextCancellation(t *testing.T) {
	mockProcessor := new(MockProcessor)
	mockLogger := new(logger.MockLogger)

	ctx, cancel := context.WithCancel(context.Background())

	requestChan := make(chan *model.MediaRequest)
	var wg sync.WaitGroup
	wg.Add(1)

	worker := NewWorker(1, mockProcessor, mockLogger)

	mockLogger.On("Info", mock.Anything, mock.Anything).Return()

	go worker.Start(ctx, &wg, requestChan)

	cancel()

	wg.Wait()

	mockLogger.AssertExpectations(t)
}
