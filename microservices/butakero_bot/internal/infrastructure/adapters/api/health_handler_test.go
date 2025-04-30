//go:build !integration

package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/adapters/api"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockDiscordChecker struct {
	mock.Mock
}

func (m *MockDiscordChecker) Check(ctx context.Context) (entity.DiscordHealth, error) {
	args := m.Called(ctx)
	return args.Get(0).(entity.DiscordHealth), args.Error(1)
}

type MockServiceBChecker struct {
	mock.Mock
}

func (m *MockServiceBChecker) Check(ctx context.Context) (entity.ServiceBHealth, error) {
	args := m.Called(ctx)
	return args.Get(0).(entity.ServiceBHealth), args.Error(1)
}

func TestHealthHandler_Handle_Success(t *testing.T) {
	// Arrange
	discordChecker := new(MockDiscordChecker)
	serviceBChecker := new(MockServiceBChecker)
	logger := new(logging.MockLogger)
	cfg := &config.Config{AppVersion: "1.0.0"}

	logger.On("With", mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()

	discordHealth := entity.DiscordHealth{
		Connected:          true,
		HeartbeatLatencyMS: 50,
		Guilds:             5,
		VoiceConnections:   2,
	}
	discordChecker.On("Check", mock.Anything).Return(discordHealth, nil)

	serviceBHealth := entity.ServiceBHealth{
		Connected: true,
		LatencyMS: 100,
		Status:    "HTTP 200",
	}
	serviceBChecker.On("Check", mock.Anything).Return(serviceBHealth, nil)

	handler := api.NewHealthHandler(discordChecker, serviceBChecker, logger, cfg)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Act
	handler.Handle(w, req)

	// Assert
	res := w.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var response entity.HealthResponse
	err := json.NewDecoder(res.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, entity.StatusOperational, response.Status)
	assert.True(t, response.Discord.Connected)
	assert.True(t, response.ServiceB.Connected)
	assert.Equal(t, "1.0.0", response.Version)

	discordChecker.AssertExpectations(t)
	serviceBChecker.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestHealthHandler_Handle_DiscordError(t *testing.T) {
	// Arrange
	discordChecker := new(MockDiscordChecker)
	serviceBChecker := new(MockServiceBChecker)
	logger := new(logging.MockLogger)
	cfg := &config.Config{AppVersion: "1.0.0"}

	logger.On("With", mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()

	discordChecker.On("Check", mock.Anything).Return(entity.DiscordHealth{}, errors.New("discord error"))

	serviceBHealth := entity.ServiceBHealth{
		Connected: true,
		LatencyMS: 100,
		Status:    "HTTP 200",
	}
	serviceBChecker.On("Check", mock.Anything).Return(serviceBHealth, nil)

	handler := api.NewHealthHandler(discordChecker, serviceBChecker, logger, cfg)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Act
	handler.Handle(w, req)

	// Assert
	res := w.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var response entity.HealthResponse
	err := json.NewDecoder(res.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, entity.StatusDegraded, response.Status)
	assert.False(t, response.Discord.Connected)
	assert.Equal(t, "discord error", response.Discord.Error)
	assert.True(t, response.ServiceB.Connected)

	discordChecker.AssertExpectations(t)
	serviceBChecker.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestHealthHandler_Handle_ServiceBError(t *testing.T) {
	// Arrange
	discordChecker := new(MockDiscordChecker)
	serviceBChecker := new(MockServiceBChecker)
	logger := new(logging.MockLogger)
	cfg := &config.Config{AppVersion: "1.0.0"}

	logger.On("With", mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()

	discordHealth := entity.DiscordHealth{
		Connected:          true,
		HeartbeatLatencyMS: 50,
		Guilds:             5,
		VoiceConnections:   2,
	}
	discordChecker.On("Check", mock.Anything).Return(discordHealth, nil)

	serviceBChecker.On("Check", mock.Anything).Return(entity.ServiceBHealth{}, errors.New("service B error"))

	handler := api.NewHealthHandler(discordChecker, serviceBChecker, logger, cfg)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Act
	handler.Handle(w, req)

	// Assert
	res := w.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var response entity.HealthResponse
	err := json.NewDecoder(res.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, entity.StatusDegraded, response.Status)
	assert.True(t, response.Discord.Connected)
	assert.False(t, response.ServiceB.Connected)
	assert.Equal(t, "service B error", response.ServiceB.Error)

	discordChecker.AssertExpectations(t)
	serviceBChecker.AssertExpectations(t)
	logger.AssertExpectations(t)
}

func TestHealthHandler_Handle_AllServicesDown(t *testing.T) {
	// Arrange
	discordChecker := new(MockDiscordChecker)
	serviceBChecker := new(MockServiceBChecker)
	logger := new(logging.MockLogger)
	cfg := &config.Config{AppVersion: "1.0.0"}

	logger.On("With", mock.Anything).Return(logger)
	logger.On("Debug", mock.Anything, mock.Anything).Return()
	logger.On("Warn", mock.Anything, mock.Anything).Return()

	discordChecker.On("Check", mock.Anything).Return(entity.DiscordHealth{}, errors.New("discord error"))
	serviceBChecker.On("Check", mock.Anything).Return(entity.ServiceBHealth{}, errors.New("service B error"))

	handler := api.NewHealthHandler(discordChecker, serviceBChecker, logger, cfg)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Act
	handler.Handle(w, req)

	// Assert
	res := w.Result()
	defer func() {
		if err := res.Body.Close(); err != nil {
			t.Fatal(err)
		}
	}()

	assert.Equal(t, http.StatusServiceUnavailable, res.StatusCode)

	var response entity.HealthResponse
	err := json.NewDecoder(res.Body).Decode(&response)
	assert.NoError(t, err)

	assert.Equal(t, entity.StatusDown, response.Status)
	assert.False(t, response.Discord.Connected)
	assert.False(t, response.ServiceB.Connected)

	discordChecker.AssertExpectations(t)
	serviceBChecker.AssertExpectations(t)
	logger.AssertExpectations(t)
}
