//go:build !integration

package processor

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestNewDownloadService(t *testing.T) {
	mockWorkerPool := new(MockDownloadSongWorkerPool)
	mockLogger := new(logger.MockLogger)

	service := NewDownloadService(mockWorkerPool, mockLogger)

	assert.NotNil(t, service)
	assert.Equal(t, mockWorkerPool, service.workerPool)
	assert.Equal(t, mockLogger, service.logger)
}

func TestDownloadService_Run_Success(t *testing.T) {
	mockWorkerPool := &MockDownloadSongWorkerPool{}
	mockLogger := new(logger.MockLogger)
	ctx := context.Background()

	service := NewDownloadService(mockWorkerPool, mockLogger)

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Once()
	mockWorkerPool.On("Start", ctx).Return(nil).Once()

	err := service.Run(ctx)

	assert.NoError(t, err)
	mockLogger.AssertExpectations(t)
	mockWorkerPool.AssertExpectations(t)
}

func TestDownloadService_Run_Error(t *testing.T) {
	mockWorkerPool := &MockDownloadSongWorkerPool{}
	mockLogger := new(logger.MockLogger)
	ctx := context.Background()
	expectedErr := errors.New("worker pool error")

	service := NewDownloadService(mockWorkerPool, mockLogger)

	mockLogger.On("With", mock.Anything, mock.Anything).Return(mockLogger)
	mockLogger.On("Info", mock.Anything, mock.Anything).Once()
	mockLogger.On("Error", mock.Anything, mock.Anything).Once()
	mockWorkerPool.On("Start", ctx).Return(expectedErr).Once()

	err := service.Run(ctx)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockLogger.AssertExpectations(t)
	mockWorkerPool.AssertExpectations(t)
}
