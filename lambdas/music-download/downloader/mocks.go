package downloader

import (
	"context"
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

type MockUploader struct {
	mock.Mock
}

func (m *MockUploader) UploadToS3(ctx context.Context, reader io.Reader, key string) error {
	args := m.Called(ctx, reader, key)
	return args.Error(0)
}

type MockCommandExecutor struct {
	mock.Mock
}

func (m *MockCommandExecutor) ExecuteCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	arguments := m.Called(ctx, name, args)
	return arguments.Get(0).([]byte), arguments.Error(1)
}
