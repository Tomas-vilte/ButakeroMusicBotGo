package middleware

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

func LoggingMiddleware(l logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Logging antes de procesar la solicitud
		l.Info("Request recibido",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("referer", c.Request.Referer()),
			zap.String("request_id", c.GetHeader("X-Request-ID")),
			zap.Int64("content_length", c.Request.ContentLength),
			zap.String("protocol", c.Request.Proto),
		)

		// Añadir el tiempo de inicio al contexto
		c.Set("startTime", start)

		c.Next()

		// Cálculo del tiempo de respuesta
		latency := time.Since(start)

		// Obtener información adicional del contexto
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Logging después de procesar la solicitud
		l.Info("Respuesta enviada",
			zap.Int("status", statusCode),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Duration("latency", latency),
			zap.Int("body_size", c.Writer.Size()),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("referer", c.Request.Referer()),
			zap.String("request_id", c.GetHeader("X-Request-ID")),
			zap.String("error", errorMessage),
		)

		// Logging de errores si los hay
		if len(c.Errors) > 0 {
			l.Error("Errores encontrados",
				zap.String("errors", c.Errors.String()),
			)
		}

		// Logging de advertencia para respuestas lentas
		if latency > time.Second*5 {
			l.Warn("Respuesta lenta detectada",
				zap.Duration("latency", latency),
				zap.String("path", path),
				zap.String("method", c.Request.Method),
			)
		}
	}
}
