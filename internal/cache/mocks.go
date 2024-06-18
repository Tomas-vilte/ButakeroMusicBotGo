package cache

import (
	"container/list"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"time"
)

// MockListInterface es un mock para ListInterface.
type MockListInterface struct {
	mock.Mock
}

func (m *MockListInterface) PushFront(v interface{}) *list.Element {
	args := m.Called(v)
	return args.Get(0).(*list.Element)
}

func (m *MockListInterface) MoveToFront(e *list.Element) {
	m.Called(e)
}

func (m *MockListInterface) Remove(e *list.Element) {
	m.Called(e)
}

func (m *MockListInterface) Back() *list.Element {
	args := m.Called()
	return args.Get(0).(*list.Element)
}

func (m *MockListInterface) Len() int {
	args := m.Called()
	return args.Int(0)
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

func (m *MockCacheMetrics) IncLatencyGet(duration time.Duration) {
	m.Called(duration)
}

func (m *MockCacheMetrics) IncLatencySet(duration time.Duration) {
	m.Called(duration)
}

// MockEntryPoolInterface es un mock para EntryPoolInterface.
type MockEntryPoolInterface struct {
	mock.Mock
}

func (m *MockEntryPoolInterface) Get() *Entry {
	args := m.Called()
	return args.Get(0).(*Entry)
}

func (m *MockEntryPoolInterface) Put(e *Entry) {
	m.Called(e)
}

// MockTimerInterface es un mock para TimerInterface.
type MockTimerInterface struct {
	mock.Mock
}

func (m *MockTimerInterface) C() <-chan time.Time {
	args := m.Called()
	return args.Get(0).(<-chan time.Time)
}

func (m *MockTimerInterface) Reset(d time.Duration) bool {
	args := m.Called(d)
	return args.Bool(0)
}

func (m *MockTimerInterface) Stop() bool {
	args := m.Called()
	return args.Bool(0)
}
