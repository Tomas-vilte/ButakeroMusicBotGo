//go:build !integration

package health_test

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/adapters/health"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	args := m.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestNewServiceBChecker_InvalidConfig(t *testing.T) {
	// Arrange
	logger := new(logging.MockLogger)

	// Act
	checker, err := health.NewServiceBChecker(nil, logger)

	// Assert
	assert.Nil(t, checker)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "La configuración del servicio no puede ser nula")
}

func TestNewServiceBChecker_InvalidURL(t *testing.T) {
	// Arrange
	logger := new(logging.MockLogger)
	config := &health.ServiceConfig{
		BaseURL: "://invalid-url",
	}

	// Act
	checker, err := health.NewServiceBChecker(config, logger)

	// Assert
	assert.Nil(t, checker)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "La URL base no es válida")
}

func TestServiceBChecker_Check_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/health", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := new(logging.MockLogger)
	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	config := &health.ServiceConfig{
		BaseURL: server.URL,
	}
	checker, _ := health.NewServiceBChecker(config, logger)

	// Act
	result, err := checker.Check(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.True(t, result.Connected)
	assert.Equal(t, "HTTP 200", result.Status)
	assert.GreaterOrEqual(t, result.LatencyMS, 0)
	logger.AssertExpectations(t)
}

func TestServiceBChecker_Check_Unhealthy(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	logger := new(logging.MockLogger)
	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()

	config := &health.ServiceConfig{
		BaseURL: server.URL,
	}
	checker, _ := health.NewServiceBChecker(config, logger)

	// Act
	result, err := checker.Check(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.False(t, result.Connected)
	assert.Equal(t, "HTTP 500", result.Status)
	assert.GreaterOrEqual(t, result.LatencyMS, 0)
	logger.AssertExpectations(t)
}

func TestServiceBChecker_Check_ConnectionError(t *testing.T) {
	// Arrange
	logger := new(logging.MockLogger)
	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	config := &health.ServiceConfig{
		BaseURL: "http://unreachable-server",
	}
	checker, _ := health.NewServiceBChecker(config, logger)

	// Act
	result, err := checker.Check(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.False(t, result.Connected)
	assert.Contains(t, result.Error, "Error de conexión")
	assert.Greater(t, result.LatencyMS, 0)
	logger.AssertExpectations(t)
}

func TestServiceBChecker_Check_Timeout(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := new(logging.MockLogger)
	logger.On("With", mock.Anything, mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

	config := &health.ServiceConfig{
		BaseURL: server.URL,
		Timeout: 100 * time.Millisecond,
	}
	checker, _ := health.NewServiceBChecker(config, logger)

	// Act
	result, err := checker.Check(context.Background())

	// Assert
	assert.NoError(t, err)
	assert.False(t, result.Connected)
	assert.Contains(t, result.Error, "context deadline exceeded")
	assert.Greater(t, result.LatencyMS, 0)
	logger.AssertExpectations(t)
}
