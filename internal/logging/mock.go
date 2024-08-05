package logging

import (
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap/zapcore"
)

// MockLogger es una implementación de mock de la interfaz Logger.
type MockLogger struct {
	mock.Mock
}

// Info registra un mensaje informativo en el mock.
func (m *MockLogger) Info(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

// Warn registra un mensaje de advertencia en el mock.
func (m *MockLogger) Warn(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

// Error registra un mensaje de error en el mock.
func (m *MockLogger) Error(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

// Debug registra un mensaje de depuración en el mock.
func (m *MockLogger) Debug(msg string, fields ...zapcore.Field) {
	m.Called(msg, fields)
}

// With añade campos adicionales a los mensajes de log en el mock.
func (m *MockLogger) With(fields ...zapcore.Field) {
	m.Called(fields)
}
