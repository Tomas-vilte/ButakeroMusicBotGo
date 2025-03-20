package middleware

import (
	"bytes"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
	"time"
)

// LoggingMiddleware registra información sobre las solicitudes HTTP.
func LoggingMiddleware(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		var requestBody []byte
		if c.Request.Body != nil && c.Request.Method != "GET" {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		log.Info("Request recibido",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
		)

		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Info("Respuesta enviada",
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("latency", latency.String()),
			zap.Int("body_size", c.Writer.Size()),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
		)

		if len(c.Errors) > 0 {
			errMsgs := make([]string, 0, len(c.Errors))
			for i, e := range c.Errors {
				errMsgs = append(errMsgs, fmt.Sprintf("Error #%02d: %s", i+1, e.Error()))
			}

			switch status {
			case http.StatusNotFound:
				log.Warn("Recurso no encontrado",
					zap.String("errors", strings.Join(errMsgs, "\n")),
					zap.Int("status", status),
				)
			case http.StatusInternalServerError:
				log.Error("Error interno del servidor",
					zap.String("errors", strings.Join(errMsgs, "\n")),
					zap.Int("status", status),
				)
			default:
				log.Warn("Errores encontrados",
					zap.String("errors", strings.Join(errMsgs, "\n")),
					zap.Int("status", status),
				)
			}
		}
	}
}

// bodyLogWriter es un writer personalizado que captura la respuesta.
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
