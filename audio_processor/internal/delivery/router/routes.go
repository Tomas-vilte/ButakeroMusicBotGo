package router

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/delivery/http/controller"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/delivery/http/middleware"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine,
	healthCheck *controller.HealthHandler,
	mediaController *controller.MediaController,
	log logger.Logger) {

	router.Use(middleware.LoggingMiddleware(log), middleware.ErrorHandlerMiddleware())

	api := router.Group("/api")
	{
		api.GET("/v1/health", healthCheck.HealthCheckHandler)
		api.GET("/v1/media", mediaController.GetMediaByID)
		api.GET("/v1/media/search", mediaController.SearchMediaByTitle)
	}
}
