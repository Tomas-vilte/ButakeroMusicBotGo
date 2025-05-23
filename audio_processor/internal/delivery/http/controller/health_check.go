package controller

import (
	"context"
	"net/http"
	"sync"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	cfg *config.Config
}

func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		cfg: cfg,
	}
}

func (h *HealthHandler) HealthCheckHandler(c *gin.Context) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make(map[string]interface{})
	allHealthy := true

	services := h.getServiceChecks()
	wg.Add(len(services))

	for name, check := range services {
		go func(name string, check func(ctx context.Context) (*api.HealthCheckMetadata, error)) {
			defer wg.Done()
			if metadata, err := check(c.Request.Context()); err != nil {
				mu.Lock()
				results[name] = map[string]interface{}{
					"status": "indisponible",
					"error":  err.Error(),
				}
				allHealthy = false
				mu.Unlock()
			} else {
				mu.Lock()
				results[name] = map[string]interface{}{
					"status":   "saludable",
					"metadata": metadata,
				}
				mu.Unlock()
			}
		}(name, check)
	}
	wg.Wait()

	status := http.StatusOK
	message := "Todos los servicios son saludables"
	if !allHealthy {
		status = http.StatusInternalServerError
		message = "Uno o mas servicios no estan saludables"
	}

	c.JSON(status, gin.H{
		"status":      message,
		"services":    results,
		"environment": h.cfg.Environment,
	})
}

func (h *HealthHandler) getServiceChecks() map[string]func(ctx context.Context) (*api.HealthCheckMetadata, error) {
	switch h.cfg.Environment {
	case "local":
		return map[string]func(ctx context.Context) (*api.HealthCheckMetadata, error){
			"mongo_db": func(ctx context.Context) (*api.HealthCheckMetadata, error) {
				metadata, err := api.CheckMongoDB(ctx, h.cfg)
				if err != nil {
					return nil, err
				}
				return &api.HealthCheckMetadata{
					Mongo: metadata,
				}, nil
			},
			"kafka": func(ctx context.Context) (*api.HealthCheckMetadata, error) {
				metadata, err := api.CheckKafka(h.cfg)
				if err != nil {
					return nil, err
				}
				return &api.HealthCheckMetadata{
					Kafka: metadata,
				}, nil
			},
		}
	case "prod":
		return map[string]func(ctx context.Context) (*api.HealthCheckMetadata, error){
			"dynamo_db": func(ctx context.Context) (*api.HealthCheckMetadata, error) {
				metadata, err := api.CheckDynamoDB(ctx, h.cfg)
				if err != nil {
					return nil, err
				}
				return &api.HealthCheckMetadata{
					DynamoDB: metadata,
				}, nil
			},
			"s3": func(ctx context.Context) (*api.HealthCheckMetadata, error) {
				metadata, err := api.CheckS3(ctx, h.cfg)
				if err != nil {
					return nil, err
				}
				return &api.HealthCheckMetadata{
					S3: metadata,
				}, nil
			},
		}
	default:
		return nil
	}
}
