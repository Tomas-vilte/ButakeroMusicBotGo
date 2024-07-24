package cache

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/types"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

func TestRedisCache_SetSong_Success(t *testing.T) {
	mockRedisClient := new(RedisClientMock)
	mockLogger := new(MockLogger)
	cache := &RedisCache{
		client:  mockRedisClient,
		logging: mockLogger,
	}

	song := &types.Song{
		Title: "Test Song",
		URL:   "https://example.com",
	}
	data, _ := json.Marshal(song)

	mockRedisClient.On("Set", mock.Anything, "test-key", string(data), 5*time.Minute).Return(redis.NewStatusResult("", nil))
	mockLogger.On("Info", "Datos guardados en cache con exito", mock.AnythingOfType("[]zapcore.Field"))

	err := cache.SetSong(context.Background(), "test-key", song)
	assert.NoError(t, err)
	mockRedisClient.AssertExpectations(t)
}

func TestRedisCache_SetSong_Error(t *testing.T) {
	mockRedisClient := new(RedisClientMock)
	mockLogger := new(MockLogger)
	cache := &RedisCache{
		client:  mockRedisClient,
		logging: mockLogger,
	}

	song := &types.Song{
		Title: "Test Song",
		URL:   "https://example.com",
	}
	data, _ := json.Marshal(song)
	mockRedisClient.On("Set", mock.Anything, "test-key", string(data), 5*time.Minute).Return(redis.NewStatusResult("", errors.New("failed to set")))
	mockLogger.On("Error", "Error al guardar en cache", mock.Anything)

	err := cache.SetSong(context.Background(), "test-key", song)
	assert.Error(t, err)
	mockRedisClient.AssertExpectations(t)
}

func TestRedisCache_GetSong(t *testing.T) {
	mockRedisClient := new(RedisClientMock)
	mockLogger := new(MockLogger)
	cache := &RedisCache{
		client:  mockRedisClient,
		logging: mockLogger,
	}

	song := &types.Song{
		Title: "Test Song",
		URL:   "https://example.com",
	}
	data, _ := json.Marshal(song)

	mockRedisClient.On("Get", mock.Anything, "test-key").Return(redis.NewStringResult(string(data), nil))
	mockLogger.On("Info", "Datos obtenidos en cache con exito", mock.Anything)

	result, err := cache.GetSong(context.Background(), "test-key")
	assert.NoError(t, err)
	assert.Equal(t, song, result)
	mockRedisClient.AssertExpectations(t)

}

func TestRedisCache_GetSong_Error(t *testing.T) {
	mockRedisClient := new(RedisClientMock)
	mockLogger := new(MockLogger)
	cache := &RedisCache{
		client:  mockRedisClient,
		logging: mockLogger,
	}

	mockRedisClient.On("Get", mock.Anything, "test-key").Return(redis.NewStringResult("", errors.New("failed to get")))
	mockLogger.On("Error", "Error al obtener del cache", mock.Anything)

	result, err := cache.GetSong(context.Background(), "test-key")
	assert.Error(t, err)
	assert.Nil(t, result)
	mockRedisClient.AssertExpectations(t)
}
