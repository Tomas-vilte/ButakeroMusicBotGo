package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"time"
)

// Logger define la interfaz para los métodos de registro de información, advertencia y error.
type (
	Logger interface {
		Info(msg string, fields ...zapcore.Field)  // Info registra un mensaje informativo.
		Warn(msg string, fields ...zapcore.Field)  // Warn registra un mensaje de advertencia.
		Error(msg string, fields ...zapcore.Field) // Error registra un mensaje de error.
		Debug(msg string, fields ...zapcore.Field) // Debug registra un mensaje de depuración.
		With(fields ...zapcore.Field) Logger       // With añade campos adicionales a los mensajes de log.
	}

	// ZapLogger es una implementación de la interfaz Logger utilizando Zap Logger.
	ZapLogger struct {
		logger *zap.Logger
	}
)

func NewProductionLogger() (*ZapLogger, error) {
	config := zap.NewProductionConfig()

	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	config.Encoding = "json"

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return &ZapLogger{logger: logger}, nil
}

func NewDevelopmentLogger() (*ZapLogger, error) {
	config := zap.NewDevelopmentConfig()

	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return &ZapLogger{logger: logger}, nil
}

// NewZapLogger crea una nueva instancia de ZapLogger.
func NewZapLogger() (*ZapLogger, error) {
	config := zap.NewProductionConfig()

	config.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     customTimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	config.Encoding = "console"

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return &ZapLogger{logger: logger}, nil
}

// Close cierra el logger.
func (l *ZapLogger) Close() error {
	err := l.logger.Sync()
	if err != nil && err.Error() != "sync /dev/stderr: invalid argument" {
		return err
	}
	return nil
}

func (l *ZapLogger) With(fields ...zapcore.Field) Logger {
	return &ZapLogger{logger: l.logger.With(fields...)}
}

// Info registra un mensaje informativo.
func (l *ZapLogger) Info(msg string, fields ...zapcore.Field) {
	l.logger.Info(msg, fields...)
}

// Warn registra un mensaje de advertencia.
func (l *ZapLogger) Warn(msg string, fields ...zapcore.Field) {
	l.logger.Warn(msg, fields...)
}

// Error registra un mensaje de error.
func (l *ZapLogger) Error(msg string, fields ...zapcore.Field) {
	l.logger.Error(msg, fields...)
}

// Debug registra un mensaje de depuración.
func (l *ZapLogger) Debug(msg string, fields ...zapcore.Field) {
	l.logger.Debug(msg, fields...)
}

// customTimeEncoder es una función para formatear la fecha y hora.
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}
