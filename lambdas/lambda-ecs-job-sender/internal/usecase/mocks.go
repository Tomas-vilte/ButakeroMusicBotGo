package usecase

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/internal/domain/entity"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockECSClient struct {
	mock.Mock
}

func (m *MockECSClient) RunTask(ctx context.Context, input *ecs.RunTaskInput) (*ecs.RunTaskOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*ecs.RunTaskOutput), args.Error(1)
}

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) GetObject(ctx context.Context, input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

type MockJobRepository struct {
	mock.Mock
}

// GetJobByS3Key simula el método GetJobByS3Key de JobRepository.
func (m *MockJobRepository) GetJobByS3Key(ctx context.Context, s3Key string) (entity.Job, error) {
	args := m.Called(ctx, s3Key)
	return args.Get(0).(entity.Job), args.Error(1)
}

// Update simula el método Update de JobRepository.
func (m *MockJobRepository) Update(ctx context.Context, job entity.Job) error {
	args := m.Called(ctx, job)
	return args.Error(0)
}

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
