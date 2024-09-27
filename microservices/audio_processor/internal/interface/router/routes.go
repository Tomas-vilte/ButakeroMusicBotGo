package router

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/interface/http/handler"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, audioHandler *handler.AudioHandler) {
	api := router.Group("/api")
	{
		api.POST("/audio/start", audioHandler.InitiateDownload)
		api.GET("/audio/status", audioHandler.GetOperationStatus)
	}
}
