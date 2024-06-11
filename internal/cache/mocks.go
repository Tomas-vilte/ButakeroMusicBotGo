package cache

import (
	"github.com/prometheus/client_golang/prometheus"
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

// MockCacheMetrics es un mock para la interfaz metrics.CacheMetrics
type MockCacheMetrics struct {
	mock.Mock
}

func (m *MockCacheMetrics) IncHits() {
	m.Called()
}

func (m *MockCacheMetrics) IncMisses() {
	m.Called()
}

func (m *MockCacheMetrics) SetCacheSize(size float64) {
	m.Called(size)
}

func (m *MockCacheMetrics) IncEvictions() {
	m.Called()
}

func (m *MockCacheMetrics) IncRequests() {
	m.Called()
}

func (m *MockCacheMetrics) IncSetOperations() {
	m.Called()
}

func (m *MockCacheMetrics) IncGetOperations() {
	m.Called()
}

func (m *MockCacheMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.Called(ch)
}

func (m *MockCacheMetrics) Collect(ch chan<- prometheus.Metric) {
	m.Called(ch)
}
