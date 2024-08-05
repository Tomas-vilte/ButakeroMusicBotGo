package cache

import (
	"container/list"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/mock"
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

// MockCacheMetrics es un mock para la interfaz metrics.CacheMetrics
type MockCacheMetrics struct {
	mock.Mock
}

func (m *MockCacheMetrics) IncHits(cacheType string) {
	m.Called(cacheType)
}

func (m *MockCacheMetrics) IncMisses(cacheType string) {
	m.Called(cacheType)
}

func (m *MockCacheMetrics) SetCacheSize(size float64) {
	m.Called(size)
}

func (m *MockCacheMetrics) IncEvictions(cacheType string) {
	m.Called()
}

func (m *MockCacheMetrics) IncRequests(cacheType string) {
	m.Called()
}

func (m *MockCacheMetrics) IncSetOperations(cacheType string) {
	m.Called()
}

func (m *MockCacheMetrics) IncGetOperations(cacheType string) {
	m.Called()
}

func (m *MockCacheMetrics) Describe(ch chan<- *prometheus.Desc) {
	m.Called(ch)
}

func (m *MockCacheMetrics) Collect(ch chan<- prometheus.Metric) {
	m.Called(ch)
}

func (m *MockCacheMetrics) IncLatencyGet(cacheType string, duration time.Duration) {
	m.Called(cacheType, duration)
}

func (m *MockCacheMetrics) IncLatencySet(cacheType string, duration time.Duration) {
	m.Called(cacheType, duration)
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
