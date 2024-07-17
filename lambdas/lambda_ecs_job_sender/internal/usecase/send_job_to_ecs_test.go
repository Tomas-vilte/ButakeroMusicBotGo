package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Tomas-vilte/GoMusicBot/lambdas/lambda-ecs-job-sender/internal/domain/entity"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSendJobToECS_Execute(t *testing.T) {
	t.Run("Successful execution", func(t *testing.T) {
		mockECS, mockS3, mockRepo, mockLogger := setupMocks()
		repo := NewSendJobsToECS(mockECS, mockS3, mockRepo, mockLogger)

		mockRepo.On("GetJobByS3Key", mock.Anything, mock.Anything).Return(entity.Job{}, nil)
		mockS3.On("GetObject", mock.Anything, mock.Anything).Return(&s3.GetObjectOutput{}, nil)
		mockECS.On("RunTask", mock.Anything, mock.Anything).Return(&ecs.RunTaskOutput{}, nil)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return(nil)
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

		err := repo.Execute(context.Background(), createTestJob())

		require.NoError(t, err)
		assertMockExpectations(t, mockECS, mockS3, mockRepo, mockLogger)
	})

	t.Run("Error checking job existence", func(t *testing.T) {
		mockECS, mockS3, mockRepo, mockLogger := setupMocks()
		repo := NewSendJobsToECS(mockECS, mockS3, mockRepo, mockLogger)

		mockRepo.On("GetJobByS3Key", mock.Anything, mock.Anything).Return(entity.Job{}, assert.AnError)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return(nil)
		mockLogger.On("Error", mock.Anything, mock.Anything).Return(nil)

		err := repo.Execute(context.Background(), createTestJob())

		require.Error(t, err)
		assert.EqualError(t, err, assert.AnError.Error())
		assertMockExpectations(t, mockECS, mockS3, mockRepo, mockLogger)
	})

	t.Run("Job already exists", func(t *testing.T) {
		mockECS, mockS3, mockRepo, mockLogger := setupMocks()
		repo := NewSendJobsToECS(mockECS, mockS3, mockRepo, mockLogger)

		mockRepo.On("GetJobByS3Key", mock.Anything, mock.Anything).Return(createTestJob(), nil)
		mockLogger.On("Info", mock.Anything, mock.Anything).Return(nil)

		err := repo.Execute(context.Background(), createTestJob())

		require.NoError(t, err)
		assertMockExpectations(t, mockECS, mockS3, mockRepo, mockLogger)
	})

	t.Run("S3 error", func(t *testing.T) {
		mockECS, mockS3, mockRepo, mockLogger := setupMocks()
		repo := NewSendJobsToECS(mockECS, mockS3, mockRepo, mockLogger)

		mockRepo.On("GetJobByS3Key", mock.Anything, mock.Anything).Return(entity.Job{}, nil)
		mockS3.On("GetObject", mock.Anything, mock.Anything).Return(&s3.GetObjectOutput{}, errors.New("s3 error"))
		mockLogger.On("Info", mock.Anything, mock.Anything).Return(nil)
		mockLogger.On("Error", mock.Anything, mock.Anything).Return(nil)

		err := repo.Execute(context.Background(), createTestJob())

		require.Error(t, err)
		assert.EqualError(t, err, "s3 error")
		assertMockExpectations(t, mockECS, mockS3, mockRepo, mockLogger)
	})

	t.Run("ECS error", func(t *testing.T) {
		mockECS, mockS3, mockRepo, mockLogger := setupMocks()
		repo := NewSendJobsToECS(mockECS, mockS3, mockRepo, mockLogger)

		mockRepo.On("GetJobByS3Key", mock.Anything, mock.Anything).Return(entity.Job{}, nil)
		mockS3.On("GetObject", mock.Anything, mock.Anything).Return(&s3.GetObjectOutput{}, nil)
		mockECS.On("RunTask", mock.Anything, mock.Anything).Return(&ecs.RunTaskOutput{}, errors.New("ECS error"))
		mockLogger.On("Info", mock.Anything, mock.Anything).Return(nil)
		mockLogger.On("Error", mock.Anything, mock.Anything).Return(nil)

		err := repo.Execute(context.Background(), createTestJob())

		require.Error(t, err)
		assert.EqualError(t, err, "ECS error")
		assertMockExpectations(t, mockECS, mockS3, mockRepo, mockLogger)
	})

	t.Run("Job update error", func(t *testing.T) {
		mockECS, mockS3, mockRepo, mockLogger := setupMocks()
		repo := NewSendJobsToECS(mockECS, mockS3, mockRepo, mockLogger)

		mockRepo.On("GetJobByS3Key", mock.Anything, mock.Anything).Return(entity.Job{}, nil)
		mockS3.On("GetObject", mock.Anything, mock.Anything).Return(&s3.GetObjectOutput{}, nil)
		mockECS.On("RunTask", mock.Anything, mock.Anything).Return(&ecs.RunTaskOutput{}, nil)
		mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update error"))
		mockLogger.On("Info", mock.Anything, mock.Anything).Return(nil)
		mockLogger.On("Error", mock.Anything, mock.Anything).Return(nil)

		err := repo.Execute(context.Background(), createTestJob())

		require.Error(t, err)
		assert.EqualError(t, err, "update error")
		assertMockExpectations(t, mockECS, mockS3, mockRepo, mockLogger)
	})
}

func setupMocks() (*MockECSClient, *MockS3Client, *MockJobRepository, *MockLogger) {
	return new(MockECSClient), new(MockS3Client), new(MockJobRepository), new(MockLogger)
}

func createTestJob() entity.Job {
	return entity.Job{
		ID:             "job1",
		KEY:            "test-key",
		S3Key:          "test-s3-key",
		BucketName:     "test-bucket",
		TaskDefinition: "test-task",
		ClusterName:    "test-cluster",
		Region:         "us-east-1",
		Status:         "",
	}
}

func assertMockExpectations(t *testing.T, mockECS *MockECSClient, mockS3 *MockS3Client, mockRepo *MockJobRepository, mockLogger *MockLogger) {
	mockRepo.AssertExpectations(t)
	mockS3.AssertExpectations(t)
	mockECS.AssertExpectations(t)
	mockLogger.AssertExpectations(t)
}
