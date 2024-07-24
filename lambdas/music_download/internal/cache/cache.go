package cache

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/types"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"time"
)

type (
	RedisClient interface {
		Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
		Get(ctx context.Context, key string) *redis.StringCmd
	}

	Cache interface {
		SetSong(ctx context.Context, key string, song *types.Song) error
		GetSong(ctx context.Context, key string) (*types.Song, error)
	}

	RedisCache struct {
		logging logging.Logger
		client  RedisClient
	}
)

func NewRedisCache(redisURL, password string, logger logging.Logger) (*RedisCache, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: password,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	})

	return &RedisCache{
		client:  client,
		logging: logger,
	}, nil
}

func (r *RedisCache) SetSong(ctx context.Context, key string, song *types.Song) error {
	data, err := json.Marshal(song)
	if err != nil {
		r.logging.Error("Error al serializar datos para cache", zap.String("key", key), zap.Error(err))
		return err
	}
	err = r.client.Set(ctx, key, string(data), 5*time.Minute).Err()
	if err != nil {
		r.logging.Error("Error al guardar en cache", zap.String("key", key), zap.Error(err))
	} else {
		r.logging.Info("Datos guardados en cache con exito", zap.String("key", key))
	}
	return err
}

func (r *RedisCache) GetSong(ctx context.Context, key string) (*types.Song, error) {
	data, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	} else if err != nil {
		r.logging.Error("Error al obtener del cache", zap.String("key", key), zap.Error(err))
		return nil, err
	}
	var song types.Song
	err = json.Unmarshal([]byte(data), &song)
	if err != nil {
		r.logging.Error("Error al deserializar cache", zap.String("key", key), zap.Error(err))
		return nil, err
	}
	r.logging.Info("Datos obtenidos en cache con exito", zap.String("key", key))
	return &song, nil
}
