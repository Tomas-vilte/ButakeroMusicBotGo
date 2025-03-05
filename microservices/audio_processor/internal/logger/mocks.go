package logger

import (
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Debug(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) With(fields ...zapcore.Field) Logger {
	args := m.Called(fields)
	return args.Get(0).(Logger)
}
