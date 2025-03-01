package service

import (
	"bytes"
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/stretchr/testify/mock"
)

type MockAudioDownloadService struct {
	mock.Mock
}

func (m *MockAudioDownloadService) DownloadAndEncode(ctx context.Context, url string) (*bytes.Buffer, error) {
	args := m.Called(ctx, url)
	return args.Get(0).(*bytes.Buffer), args.Error(1)
}

type MockAudioStorageService struct {
	mock.Mock
}

func (m *MockAudioStorageService) StoreAudio(ctx context.Context, buffer *bytes.Buffer, metadata *model.Metadata) (*model.FileData, error) {
	args := m.Called(ctx, buffer, metadata)
	return args.Get(0).(*model.FileData), args.Error(1)
}

type MockOperationsManager struct {
	mock.Mock
}

func (m *MockOperationsManager) HandleOperationSuccess(ctx context.Context, operationID string, metadata *model.Metadata, fileData *model.FileData) error {
	args := m.Called(ctx, operationID, metadata, fileData)
	return args.Error(0)
}

type MockMessagingManager struct {
	mock.Mock
}

func (m *MockMessagingManager) SendProcessingMessage(ctx context.Context, operationID string, status string, metadata *model.Metadata, attempts int) error {
	args := m.Called(ctx, operationID, status, metadata, attempts)
	return args.Error(0)
}

type MockErrorManagement struct {
	mock.Mock
}

func (m *MockErrorManagement) HandleProcessingError(ctx context.Context, operationID string, metadata *model.Metadata, stage string, attempts int, err error) error {
	args := m.Called(ctx, operationID, metadata, stage, attempts, err)
	return args.Error(0)
}
