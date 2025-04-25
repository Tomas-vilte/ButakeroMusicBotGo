package voice

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"io"
	"sync/atomic"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/decoder"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

var (
	ErrNoVoiceConnection = errors.New("no hay conexión de voz activa")
	ErrSendTimeout       = errors.New("tiempo de espera agotado al enviar el frame de audio")
)

type DiscordVoiceSession struct {
	session     *discordgo.Session
	guildID     string
	channelID   string
	vc          *discordgo.VoiceConnection
	logger      logging.Logger
	isPaused    atomic.Bool
	sendTimeout time.Duration
	maxRetries  int
	retryDelay  time.Duration
}

type SessionOption func(*DiscordVoiceSession)

// WithSendTimeout configura el timeout para enviar frames de audio
func WithSendTimeout(timeout time.Duration) SessionOption {
	return func(d *DiscordVoiceSession) {
		d.sendTimeout = timeout
	}
}

// WithRetryConfig configura los parámetros de reintento
func WithRetryConfig(maxRetries int, retryDelay time.Duration) SessionOption {
	return func(d *DiscordVoiceSession) {
		d.maxRetries = maxRetries
		d.retryDelay = retryDelay
	}
}

func NewDiscordVoiceSession(s *discordgo.Session, guildID string, logger logging.Logger, opts ...SessionOption) *DiscordVoiceSession {
	session := &DiscordVoiceSession{
		session:     s,
		guildID:     guildID,
		logger:      logger,
		sendTimeout: 3 * time.Second,
		maxRetries:  3,
		retryDelay:  2 * time.Second,
	}

	for _, opt := range opts {
		opt(session)
	}

	return session
}

// JoinVoiceChannel conecta la sesión a un canal de voz específico usando channelID con reintentos
func (d *DiscordVoiceSession) JoinVoiceChannel(ctx context.Context, channelID string) error {
	logger := d.logger.With(
		zap.String("component", "DiscordVoiceSession"),
		zap.String("method", "JoinVoiceChannel"),
		zap.String("channelID", channelID),
		zap.String("guild_id", d.guildID),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	d.channelID = channelID
	var err error

	for attempt := 0; attempt <= d.maxRetries; attempt++ {
		if attempt > 0 {
			logger.Info("Reintentando conexión al canal de voz",
				zap.Int("retry_attempt", attempt),
				zap.Duration("retry_delay", d.retryDelay))
			select {
			case <-time.After(d.retryDelay):
			case <-ctx.Done():
				return fmt.Errorf("contexto cancelado durante reintento: %w", ctx.Err())
			}
		}

		logger.Debug("Conectando al canal de voz", zap.Int("attempt", attempt+1))

		d.vc, err = d.session.ChannelVoiceJoin(d.guildID, channelID, false, true)
		if err == nil {
			logger.Debug("Conexión de voz establecida exitosamente",
				zap.Int("attempts_needed", attempt+1))
			return nil
		}

		logger.Warn("Intento de conexión fallido",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", d.maxRetries))
	}

	logger.Error("Falló la conexión al canal de voz después de múltiples intentos",
		zap.Error(err),
		zap.Int("max_retries", d.maxRetries))
	return fmt.Errorf("error al conectar después de %d intentos: %w", d.maxRetries+1, err)
}

// ReconnectVoiceChannel intenta reconectar a la sesión de voz actual
func (d *DiscordVoiceSession) ReconnectVoiceChannel(ctx context.Context) error {
	logger := d.logger.With(
		zap.String("component", "DiscordVoiceSession"),
		zap.String("method", "ReconnectVoiceChannel"),
		zap.String("guild_id", d.guildID),
		zap.String("channelID", d.channelID),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	if d.channelID == "" {
		logger.Warn("Intento de reconexión sin canal previo")
		return errors.New("no hay canal previo para reconectar")
	}

	if d.vc != nil {
		logger.Debug("Cerrando conexión anterior antes de reconectar")
		if err := d.vc.Disconnect(); err != nil {
			logger.Warn("Error al cerrar conexión anterior", zap.Error(err))
		}
		d.vc = nil
	}

	logger.Info("Intentando reconexión al canal de voz")
	return d.JoinVoiceChannel(ctx, d.channelID)
}

// SendAudio manda frames de audio a la conexión de voz de Discord con manejo de reconexión
func (d *DiscordVoiceSession) SendAudio(ctx context.Context, reader io.ReadCloser) error {
	logger := d.logger.With(
		zap.String("component", "DiscordVoiceSession"),
		zap.String("method", "SendAudio"),
		zap.String("guild_id", d.guildID),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	if d.vc == nil {
		logger.Error("Intento de enviar audio sin conexión de voz")
		return ErrNoVoiceConnection
	}

	defer func() {
		if err := d.vc.Speaking(false); err != nil {
			logger.Error("Error al dejar de hablar",
				zap.Error(err))
		}
	}()

	logger.Debug("Iniciando transmisión de audio")

	err := d.retryOperation(ctx, func() error {
		return d.vc.Speaking(true)
	}, "iniciar Speaking")

	if err != nil {
		logger.Error("Error persistente al empezar a hablar", zap.Error(err))
		return err
	}

	decoderAudio := decoder.NewBufferedOpusDecoder(reader)
	frameCount := 0
	consecutiveErrors := 0
	maxConsecutiveErrors := 5

	for {
		select {
		case <-ctx.Done():
			logger.Debug("Transmisión cancelada por contexto",
				zap.String("reason", ctx.Err().Error()),
				zap.Int("frames_sent", frameCount))
			return ctx.Err()
		default:
		}

		if d.isPaused.Load() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		frame, err := decoderAudio.OpusFrame()
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Debug("Transmisión de audio completada",
					zap.Int("total_frames", frameCount))
				return nil
			}
			logger.Error("Error decodificando frame de audio",
				zap.Error(err),
				zap.Int("frames_sent", frameCount))
			return err
		}

		sendSuccessful := false
		for retryCount := 0; retryCount <= d.maxRetries && !sendSuccessful; retryCount++ {
			if retryCount > 0 {
				logger.Debug("Reintentando envío de frame",
					zap.Int("retry", retryCount),
					zap.Int("frame", frameCount))
				time.Sleep(50 * time.Millisecond)
			}

			select {
			case d.vc.OpusSend <- frame:
				frameCount++
				sendSuccessful = true
				consecutiveErrors = 0
			case <-ctx.Done():
				logger.Debug("Transmisión interrumpida durante envío",
					zap.String("reason", ctx.Err().Error()),
					zap.Int("frames_sent", frameCount))
				return ctx.Err()
			case <-time.After(d.sendTimeout):
				if retryCount == d.maxRetries {
					consecutiveErrors++
					logger.Warn("Timeout persistente al enviar frame de audio",
						zap.Duration("timeout", d.sendTimeout),
						zap.Int("frames_sent", frameCount),
						zap.Int("consecutive_errors", consecutiveErrors))

					if consecutiveErrors >= maxConsecutiveErrors {
						logger.Warn("Demasiados errores consecutivos, intentando reconexión",
							zap.Int("consecutive_errors", consecutiveErrors))

						reconErr := d.ReconnectVoiceChannel(ctx)
						if reconErr != nil {
							logger.Error("Falló la reconexión después de errores consecutivos",
								zap.Error(reconErr))
							return fmt.Errorf("error de envío persistente y reconexión fallida: %w", reconErr)
						}

						consecutiveErrors = 0

						if spkErr := d.vc.Speaking(true); spkErr != nil {
							logger.Error("Error al reanudar Speaking después de reconexión",
								zap.Error(spkErr))
							return spkErr
						}
					}
				}
			}
		}

		if !sendSuccessful {
			logger.Error("No se pudo enviar el frame después de todos los reintentos",
				zap.Int("frames_sent", frameCount),
				zap.Int("max_retries", d.maxRetries))
			return ErrSendTimeout
		}
	}
}

// retryOperation ejecuta una operación con reintentos
func (d *DiscordVoiceSession) retryOperation(ctx context.Context, operation func() error, operationName string) error {
	logger := d.logger.With(
		zap.String("component", "DiscordVoiceSession"),
		zap.String("method", "retryOperation"),
		zap.String("operation", operationName),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	var lastErr error
	for attempt := 0; attempt <= d.maxRetries; attempt++ {
		if attempt > 0 {
			logger.Debug("Reintentando operación",
				zap.String("operation", operationName),
				zap.Int("attempt", attempt))
			select {
			case <-time.After(d.retryDelay):
			case <-ctx.Done():
				return fmt.Errorf("contexto cancelado durante reintento de %s: %w", operationName, ctx.Err())
			}
		}

		if err := operation(); err == nil {
			if attempt > 0 {
				logger.Debug("Operación exitosa después de reintentos",
					zap.String("operation", operationName),
					zap.Int("attempts", attempt+1))
			}
			return nil
		} else {
			lastErr = err
			logger.Warn("Fallo en operación",
				zap.String("operation", operationName),
				zap.Error(err),
				zap.Int("attempt", attempt+1),
				zap.Int("max_retries", d.maxRetries))
		}
	}

	logger.Error("Operación fallida después de reintentos",
		zap.String("operation", operationName),
		zap.Error(lastErr),
		zap.Int("max_retries", d.maxRetries))
	return fmt.Errorf("%s: %w (después de %d intentos)", operationName, lastErr, d.maxRetries+1)
}

// Pause pausa la reproducción de audio
func (d *DiscordVoiceSession) Pause() {
	if !d.isPaused.Swap(true) {
		d.logger.Debug("Reproducción pausada")
	}
}

// Resume reanuda la reproducción de audio
func (d *DiscordVoiceSession) Resume() {
	if d.isPaused.Swap(false) {
		d.logger.Debug("Reproducción reanudada")
	}
}

// LeaveVoiceChannel desconecta del canal de voz
func (d *DiscordVoiceSession) LeaveVoiceChannel(ctx context.Context) error {
	logger := d.logger.With(
		zap.String("component", "DiscordVoiceSession"),
		zap.String("method", "LeaveVoiceChannel"),
		zap.String("guild_id", d.guildID),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)
	if d.vc == nil {
		logger.Debug("Intento de desconexión sin conexión activa")
		return nil
	}

	logger.Debug("Desconectando del canal de voz")
	err := d.vc.Disconnect()
	d.vc = nil

	if err != nil {
		logger.Error("Error al desconectar del canal de voz",
			zap.Error(err))
		return err
	}

	logger.Debug("Desconexión exitosa")
	return nil
}

// IsConnected comprueba si la sesión de voz está conectada
func (d *DiscordVoiceSession) IsConnected() bool {
	return d.vc != nil && d.vc.Ready
}

// SetSendTimeout configura el timeout para enviar frames de audio
func (d *DiscordVoiceSession) SetSendTimeout(timeout time.Duration) {
	d.logger.Info("Actualizando timeout de envío",
		zap.String("component", "DiscordVoiceSession"),
		zap.String("guild_id", d.guildID),
		zap.String("method", "SetSendTimeout"),
		zap.Duration("old_timeout", d.sendTimeout),
		zap.Duration("new_timeout", timeout))
	d.sendTimeout = timeout
}
