package handler

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
	"github.com/gin-gonic/gin"
	"net/http"
)

type HealthHandler struct {
	youtubeAPIKey string
}

func NewHealthHandler(youtubeAPIKey string) *HealthHandler {
	return &HealthHandler{
		youtubeAPIKey: youtubeAPIKey,
	}
}

func (h *HealthHandler) HealthCheckHandler(c *gin.Context) {
	services := map[string]func() error{
		"DynamoDB": api.CheckDynamoDB,
		"S3":       api.CheckS3,
		"YouTube API": func() error {
			return api.CheckYouTube(h.youtubeAPIKey)
		},
	}
	results := make(map[string]string)
	allHealthy := true

	for name, check := range services {
		if err := check(); err != nil {
			results[name] = "indisponible" + err.Error()
			allHealthy = false
		} else {
			results[name] = "saludable"
		}
	}

	status := http.StatusOK
	message := "Todos los servicios son saludables."
	if !allHealthy {
		status = http.StatusInternalServerError
		message = "Uno o mas servicios no estan saludables"
	}

	c.JSON(status, gin.H{
		"status":   message,
		"services": results,
	})
}
