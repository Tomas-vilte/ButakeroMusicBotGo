package router

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/interface/http/handler"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, audioHandler *handler.AudioHandler, healthCheck *handler.HealthHandler, log logger.Logger) {
	router.Use(middleware.LoggingMiddleware(log))
	api := router.Group("/api")
	{
		api.POST("/audio/start", audioHandler.InitiateDownload)
		api.GET("/audio/status", audioHandler.GetOperationStatus)
		api.GET("/health", healthCheck.HealthCheckHandler)
	}
}
