package middleware

import (
	"errors"
	errApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/gin-gonic/gin"
	"strings"
)

// ErrorHandlerMiddleware maneja errores y proporciona respuestas de error uniformes.
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			var appErr *errApp.AppError
			var aerr *errApp.AppError
			if errors.As(err, &aerr) {
				appErr = aerr
			}

			errorResponse := gin.H{
				"error": gin.H{
					"code":    appErr.Code,
					"message": appErr.Message,
				},
				"success": false,
			}

			if gin.Mode() != gin.ReleaseMode && appErr.Err != nil {
				details := strings.Split(appErr.Err.Error(), "\n")[0]
				errorResponse["error"].(gin.H)["details"] = details
			}

			statusCode := appErr.StatusCode()
			c.JSON(statusCode, errorResponse)

		}
	}
}
