package lambda

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/internal/domain/entity"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) With(fields ...zap.Field) {
	m.Called(fields)
}

type MockSendJobToECS struct {
	mock.Mock
}

func (m *MockSendJobToECS) Execute(ctx context.Context, job entity.Job) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}
