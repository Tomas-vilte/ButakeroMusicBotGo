package middleware

import (
	"errors"
	errApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

// ErrorHandlerMiddleware maneja errores y proporciona respuestas de error uniformes.
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			var appErr *errApp.AppError
			if errors.As(err, &appErr) {
				errorResponse := gin.H{
					"error": gin.H{
						"code":    appErr.Code,
						"message": appErr.Message,
					},
					"success": false,
				}

				if gin.Mode() != gin.ReleaseMode && appErr.Err != nil {
					errorResponse["error"].(gin.H)["details"] = appErr.Err.Error()
				}

				c.JSON(appErr.StatusCode(), errorResponse)
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "internal_error",
					"message": "Ocurrió un error interno en el servidor.",
				},
				"success": false,
			})
		}
	}
}
