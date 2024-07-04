package uploader

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
	"io"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) With(fields ...zapcore.Field) {
	m.Called(fields)
}

type MockS3Downloader struct {
	mock.Mock
}

type MockUploader struct {
	mock.Mock
}

func (m *MockUploader) UploadToS3(ctx context.Context, audioData io.Reader, key string) error {
	args := m.Called(ctx, audioData, key)
	return args.Error(0)
}

type MockS3Uploader struct {
	mock.Mock
}

func (m *MockS3Uploader) UploadWithContext(ctx aws.Context, input *s3manager.UploadInput, opts ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*s3manager.UploadOutput), args.Error(1)
}
