package router

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/interface/http/handler"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine,
	audioHandler *handler.AudioHandler,
	operationHandler *handler.OperationHandler,
	healthCheck *handler.HealthHandler,
	log logger.Logger) {

	router.Use(middleware.LoggingMiddleware(log))
	router.Use(middleware.ErrorHandlerMiddleware(log))

	api := router.Group("/api")
	{
		api.POST("/v1/audio/start", audioHandler.InitiateDownload)
		api.GET("/v1/operations/status", operationHandler.GetOperationStatus)
		api.GET("/v1/health", healthCheck.HealthCheckHandler)
	}
}
