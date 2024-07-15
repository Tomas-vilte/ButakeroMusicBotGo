package uploader

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

type MockS3Uploader struct {
	mock.Mock
}

func (m *MockS3Uploader) UploadWithContext(ctx aws.Context, input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	args := m.Called(ctx, input, options)
	return args.Get(0).(*s3manager.UploadOutput), args.Error(1)
}

type MockS3Client struct {
	mock.Mock
}

func (m *MockS3Client) GetObjectWithContext(ctx aws.Context, input *s3.GetObjectInput, opts ...request.Option) (*s3.GetObjectOutput, error) {
	args := m.Called(ctx, input, opts)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

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
