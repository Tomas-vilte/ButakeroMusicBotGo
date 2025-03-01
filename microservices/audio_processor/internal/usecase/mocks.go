package usecase

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/stretchr/testify/mock"
)

type MockOperationRepository struct {
	mock.Mock
}

func (m *MockOperationRepository) SaveOperationsResult(ctx context.Context, result *model.OperationResult) error {
	args := m.Called(ctx, result)
	return args.Error(0)
}

func (m *MockOperationRepository) GetOperationResult(ctx context.Context, id, songID string) (*model.OperationResult, error) {
	args := m.Called(ctx, id, songID)
	return args.Get(0).(*model.OperationResult), args.Error(1)
}

func (m *MockOperationRepository) DeleteOperationResult(ctx context.Context, id, songID string) error {
	args := m.Called(ctx, id, songID)
	return args.Error(0)
}

func (m *MockOperationRepository) UpdateOperationStatus(ctx context.Context, operationID, songID, status string) error {
	args := m.Called(ctx, operationID, songID, status)
	return args.Error(0)
}

func (m *MockOperationRepository) UpdateOperationResult(ctx context.Context, operationID string, operationResult *model.OperationResult) error {
	args := m.Called(ctx, operationID, operationResult)
	return args.Error(0)
}
