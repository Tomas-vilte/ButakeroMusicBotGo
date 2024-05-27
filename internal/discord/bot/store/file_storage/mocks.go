package file_storage

import (
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

type MockStatePersistent struct {
	mock.Mock
}

func (m *MockStatePersistent) ReadState(filepath string) (*FileState, error) {
	args := m.Called(filepath)
	return args.Get(0).(*FileState), args.Error(1)
}

func (m *MockStatePersistent) WriteState(filepath string, state *FileState) error {
	args := m.Called(filepath, state)
	return args.Error(0)
}
