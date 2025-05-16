package voice

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/interfaces"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/trace"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/decoder"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

var (
	ErrNoVoiceConnection = errors.New("no hay conexión de voz activa")
)

var _ interfaces.VoiceConnection = (*DiscordVoiceSession)(nil)

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
	vcMu        sync.RWMutex
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

func (d *DiscordVoiceSession) getValidVc() (*discordgo.VoiceConnection, error) {
	d.vcMu.RLock()
	defer d.vcMu.RUnlock()
	if d.vc == nil {
		return nil, ErrNoVoiceConnection
	}
	return d.vc, nil
}

func (d *DiscordVoiceSession) JoinVoiceChannel(ctx context.Context, targetChannelID string) error {
	logger := d.logger.With(
		zap.String("component", "DiscordVoiceSession"),
		zap.String("method", "JoinVoiceChannel"),
		zap.String("targetChannelID", targetChannelID),
		zap.String("guild_id", d.guildID),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	if targetChannelID == "" {
		return errors.New("ID de canal de voz vacío")
	}

	d.vcMu.Lock()
	defer d.vcMu.Unlock()

	if d.vc != nil && d.channelID == targetChannelID && d.vc.Ready {
		logger.Debug("Ya conectado al canal de voz solicitado", zap.String("channelID", targetChannelID))
		return nil
	}

	if d.vc != nil {
		logger.Info("Desconectando de un canal de voz previo o no listo", zap.String("previousChannelID", d.channelID))
		if err := d.vc.Disconnect(); err != nil {
			logger.Warn("Error al desconectar de la conexión de voz previa", zap.Error(err))
		}
		d.vc = nil
	}

	d.channelID = targetChannelID
	var err error
	var newVc *discordgo.VoiceConnection

	for attempt := 0; attempt <= d.maxRetries; attempt++ {
		if attempt > 0 {
			logger.Info("Reintentando conexión al canal de voz",
				zap.Int("retry_attempt", attempt),
				zap.Duration("retry_delay", d.retryDelay))
			select {
			case <-time.After(d.retryDelay):
			case <-ctx.Done():
				return fmt.Errorf("contexto cancelado durante reintento de JoinVoiceChannel: %w", ctx.Err())
			}
		}

		logger.Debug("Conectando al canal de voz", zap.Int("attempt", attempt+1))
		newVc, err = d.session.ChannelVoiceJoin(d.guildID, targetChannelID, false, true)
		if err == nil {
			logger.Debug("Conexión de voz establecida exitosamente", zap.Int("attempts_needed", attempt+1))
			d.vc = newVc
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
	d.vc = nil
	return fmt.Errorf("error al conectar después de %d intentos en JoinVoiceChannel: %w", d.maxRetries+1, err)
}

func (d *DiscordVoiceSession) ReconnectVoiceChannel(ctx context.Context) error {
	logger := d.logger.With(
		zap.String("component", "DiscordVoiceSession"),
		zap.String("method", "ReconnectVoiceChannel"),
		zap.String("guild_id", d.guildID),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	d.vcMu.RLock()
	currentStoredChannelID := d.channelID
	d.vcMu.RUnlock()

	if currentStoredChannelID == "" {
		logger.Warn("Intento de reconexión sin canal previo almacenado")
		return errors.New("no hay canal previo para reconectar")
	}

	logger.Info("Intentando reconexión al canal de voz", zap.String("channelID", currentStoredChannelID))
	return d.JoinVoiceChannel(ctx, currentStoredChannelID)
}

func (d *DiscordVoiceSession) SendAudio(ctx context.Context, reader io.ReadCloser) error {
	logger := d.logger.With(
		zap.String("component", "DiscordVoiceSession"),
		zap.String("method", "SendAudio"),
		zap.String("guild_id", d.guildID),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	defer func() {
		vc, _ := d.getValidVc()
		if vc != nil {
			if err := vc.Speaking(false); err != nil {
				logger.Error("Error al ejecutar Speaking(false) en defer", zap.Error(err))
			}
		}
	}()

	err := d.retryOperation(ctx, func() error {
		vc, err := d.getValidVc()
		if err != nil {
			return err
		}
		return vc.Speaking(true)
	}, "iniciar Speaking")

	if err != nil {
		logger.Error("Error persistente al empezar a hablar (Speaking(true))", zap.Error(err))
		return err
	}

	logger.Debug("Iniciando transmisión de audio")
	decoderAudio := decoder.NewBufferedOpusDecoder(reader)
	frameCount := 0
	consecutiveSendErrors := 0
	maxConsecutiveSendErrors := 5

	for {
		select {
		case <-ctx.Done():
			logger.Debug("Transmisión cancelada por contexto", zap.Error(ctx.Err()), zap.Int("frames_sent", frameCount))
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
				logger.Debug("Transmisión de audio completada (EOF)", zap.Int("total_frames", frameCount))
				return nil
			}
			logger.Error("Error decodificando frame de audio", zap.Error(err), zap.Int("frames_sent", frameCount))
			return err
		}

		sendSuccessful := false
		for retryCount := 0; retryCount <= d.maxRetries && !sendSuccessful; retryCount++ {
			if retryCount > 0 {
				logger.Debug("Reintentando envío de frame", zap.Int("retry", retryCount), zap.Int("frame", frameCount))
				time.Sleep(50 * time.Millisecond)
			}

			currentVc, vcErr := d.getValidVc()
			if vcErr != nil {
				logger.Error("No hay conexión de voz válida para enviar frame, intentando reconexión", zap.Error(vcErr), zap.Int("frame", frameCount))
				if reconErr := d.ReconnectVoiceChannel(ctx); reconErr != nil {
					logger.Error("Falló la reconexión durante el envío de frame", zap.Error(reconErr))
					return fmt.Errorf("conexión perdida y reconexión fallida durante envío: %w", reconErr)
				}
				if spkErr := d.retryOperation(ctx, func() error {
					vcPostRecon, errPostRecon := d.getValidVc()
					if errPostRecon != nil {
						return errPostRecon
					}
					return vcPostRecon.Speaking(true)
				}, "reanudar Speaking post-reconexión en bucle de envío"); spkErr != nil {
					logger.Error("Error al reanudar Speaking después de reconexión forzada en envío", zap.Error(spkErr))
					return spkErr
				}
				currentVc, vcErr = d.getValidVc()
				if vcErr != nil {
					logger.Error("Incapaz de obtener VC válido incluso después de reconexión forzada", zap.Error(vcErr))
					return fmt.Errorf("VC inválido después de reconexión forzada: %w", vcErr)
				}
			}

			select {
			case currentVc.OpusSend <- frame:
				frameCount++
				sendSuccessful = true
				consecutiveSendErrors = 0
			case <-ctx.Done():
				logger.Debug("Transmisión interrumpida durante envío de frame (contexto)", zap.Error(ctx.Err()), zap.Int("frames_sent", frameCount))
				return ctx.Err()
			case <-time.After(d.sendTimeout):
				logger.Warn("Timeout al enviar frame de audio",
					zap.Duration("timeout", d.sendTimeout),
					zap.Int("frames_sent", frameCount),
					zap.Int("retry_attempt", retryCount))
				if retryCount == d.maxRetries {
					consecutiveSendErrors++
				}
			}
		}

		if !sendSuccessful {
			logger.Error("No se pudo enviar el frame después de todos los reintentos",
				zap.Int("frames_sent", frameCount),
				zap.Int("max_retries", d.maxRetries))
			consecutiveSendErrors = maxConsecutiveSendErrors
		}

		if consecutiveSendErrors >= maxConsecutiveSendErrors {
			logger.Warn("Demasiados errores de envío consecutivos, intentando reconexión",
				zap.Int("consecutive_errors", consecutiveSendErrors))

			reconErr := d.ReconnectVoiceChannel(ctx)
			if reconErr != nil {
				logger.Error("Falló la reconexión después de errores de envío consecutivos", zap.Error(reconErr))
				return fmt.Errorf("error de envío persistente y reconexión fallida: %w", reconErr)
			}
			consecutiveSendErrors = 0

			if spkErr := d.retryOperation(ctx, func() error {
				vc, err := d.getValidVc()
				if err != nil {
					return err
				}
				return vc.Speaking(true)
			}, "reanudar Speaking post-reconexión por errores de envío"); spkErr != nil {
				logger.Error("Error al reanudar Speaking después de reconexión (errores de envío)", zap.Error(spkErr))
				return spkErr
			}
			logger.Info("Reconexión y Speaking(true) reanudado exitosamente tras errores de envío")
		}
	}
}

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
			logger.Debug("Reintentando operación", zap.Int("attempt", attempt))
			select {
			case <-time.After(d.retryDelay):
			case <-ctx.Done():
				return fmt.Errorf("contexto cancelado durante reintento de %s: %w", operationName, ctx.Err())
			}
		}

		if err := operation(); err == nil {
			if attempt > 0 {
				logger.Debug("Operación exitosa después de reintentos", zap.Int("attempts_needed", attempt+1))
			}
			return nil
		} else {
			lastErr = err
			if errors.Is(err, ErrNoVoiceConnection) && (operationName == "iniciar Speaking" || operationName == "reanudar Speaking post-reconexión") {
				logger.Error("Fallo crítico en operación debido a ErrNoVoiceConnection, no se reintentará esta instancia", zap.Error(err))
				return err
			}
			logger.Warn("Fallo en operación, reintentando", zap.Error(err), zap.Int("attempt", attempt+1))
		}
	}

	logger.Error("Operación fallida después de todos los reintentos", zap.Error(lastErr), zap.Int("max_retries", d.maxRetries))
	return fmt.Errorf("%s: %w (después de %d intentos)", operationName, lastErr, d.maxRetries+1)
}

func (d *DiscordVoiceSession) Pause() {
	if !d.isPaused.Swap(true) {
		d.logger.Debug("Reproducción pausada",
			zap.String("component", "DiscordVoiceSession"),
			zap.String("guild_id", d.guildID))
	}
}

func (d *DiscordVoiceSession) Resume() {
	if d.isPaused.Swap(false) {
		d.logger.Debug("Reproducción reanudada",
			zap.String("component", "DiscordVoiceSession"),
			zap.String("guild_id", d.guildID))
	}
}

func (d *DiscordVoiceSession) LeaveVoiceChannel(ctx context.Context) error {
	logger := d.logger.With(
		zap.String("component", "DiscordVoiceSession"),
		zap.String("method", "LeaveVoiceChannel"),
		zap.String("guild_id", d.guildID),
		zap.String("trace_id", trace.GetTraceID(ctx)),
	)

	d.vcMu.Lock()
	defer d.vcMu.Unlock()

	if d.vc == nil {
		logger.Debug("Intento de desconexión sin conexión activa, no se hace nada")
		return nil
	}

	logger.Debug("Desconectando del canal de voz...")
	err := d.vc.Disconnect()
	d.vc = nil
	d.channelID = ""

	if err != nil {
		logger.Error("Error al desconectar del canal de voz", zap.Error(err))
		return err
	}

	logger.Debug("Desconexión exitosa del canal de voz")
	return nil
}

func (d *DiscordVoiceSession) IsConnected() bool {
	d.vcMu.RLock()
	defer d.vcMu.RUnlock()
	return d.vc != nil && d.vc.Ready
}

func (d *DiscordVoiceSession) SetSendTimeout(timeout time.Duration) {
	d.logger.Info("Actualizando timeout de envío",
		zap.String("component", "DiscordVoiceSession"),
		zap.String("guild_id", d.guildID),
		zap.String("method", "SetSendTimeout"),
		zap.Duration("old_timeout", d.sendTimeout),
		zap.Duration("new_timeout", timeout))
	d.sendTimeout = timeout
}
