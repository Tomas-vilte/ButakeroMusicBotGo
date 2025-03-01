package middleware

import (
	"errors"
	errorsApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

func ErrorHandlerMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			var appErr *errorsApp.AppError
			if errors.As(err, &appErr) {
				if appErr.StatusCode() >= http.StatusInternalServerError {
					log.Error("Error del servidor",
						zap.String("code", appErr.Code),
						zap.Error(err))
				} else {
					log.Info("Error del cliente",
						zap.String("code", appErr.Code),
						zap.Error(err))
				}

				c.JSON(appErr.StatusCode(), gin.H{
					"code":    appErr.Code,
					"message": appErr.Message,
				})
				return
			}

			log.Error("Error interno no manejado", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "internal_error",
				"message": "Error interno del servidor",
			})
		}
	}
}
