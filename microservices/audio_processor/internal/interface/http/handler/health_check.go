package handler

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
)

type HealthHandler struct {
	cfg config.Config
}

func NewHealthHandler(cfg config.Config) *HealthHandler {
	return &HealthHandler{
		cfg: cfg,
	}
}

func (h *HealthHandler) HealthCheckHandler(c *gin.Context) {
	var wg sync.WaitGroup
	wg.Add(2)

	var mu sync.Mutex

	results := make(map[string]string)
	allHealthy := true

	services := map[string]func(ctx context.Context) error{
		"DynamoDB": func(ctx context.Context) error {
			defer wg.Done()
			return api.CheckDynamoDB(ctx, h.cfg)
		},
		"S3": func(ctx context.Context) error {
			defer wg.Done()
			return api.CheckS3(ctx, h.cfg)
		},
	}

	for name, check := range services {
		go func(name string, check func(ctx context.Context) error) {
			if err := check(c.Request.Context()); err != nil {
				mu.Lock()
				results[name] = "indisponible: " + err.Error()
				allHealthy = false
				mu.Unlock()
			} else {
				mu.Lock()
				results[name] = "saludable"
				mu.Unlock()
			}
		}(name, check)
	}

	wg.Wait()

	status := http.StatusOK
	message := "Todos los servicios son saludables."
	if !allHealthy {
		status = http.StatusInternalServerError
		message = "Uno o más servicios no están saludables"
	}

	c.JSON(status, gin.H{
		"status":   message,
		"services": results,
	})
}
