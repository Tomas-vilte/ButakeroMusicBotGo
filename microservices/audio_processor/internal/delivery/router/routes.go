package router

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/delivery/http/handler"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/delivery/http/middleware"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine,
	operationHandler *handler.OperationHandler,
	healthCheck *handler.HealthHandler,
	log logger.Logger) {

	router.Use(middleware.LoggingMiddleware(log), middleware.ErrorHandlerMiddleware())

	api := router.Group("/api")
	{
		api.GET("/v1/operations/status", operationHandler.GetOperationStatus)
		api.GET("/v1/health", healthCheck.HealthCheckHandler)
	}
}
