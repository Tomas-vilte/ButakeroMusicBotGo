package voice

import (
	"context"
	"errors"
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
	vc          *discordgo.VoiceConnection
	logger      logging.Logger
	isPaused    atomic.Bool
	sendTimeout time.Duration
}

func NewDiscordVoiceSession(s *discordgo.Session, guildID string, logger logging.Logger) *DiscordVoiceSession {
	return &DiscordVoiceSession{
		session:     s,
		guildID:     guildID,
		logger:      logger,
		sendTimeout: 3 * time.Second,
	}
}

// JoinVoiceChannel conecta la sesión a un canal de voz específico usando channelID
func (d *DiscordVoiceSession) JoinVoiceChannel(channelID string) error {
	logger := d.logger.With(
		zap.String("component", "DiscordVoiceSession"),
		zap.String("method", "JoinVoiceChannel"),
		zap.String("channelID", channelID),
		zap.String("guild_id", d.guildID),
	)
	logger.Debug("Conectando al canal de voz", zap.String("channelID", channelID))
	vc, err := d.session.ChannelVoiceJoin(d.guildID, channelID, false, true)
	if err != nil {
		logger.Error("Falló la conexión al canal de voz", zap.Error(err))
		return err
	}
	d.vc = vc
	logger.Debug("Conexión de voz establecida exitosamente")
	return nil
}

// SendAudio manda frames de audio a la conexión de voz de Discord
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
		if err := reader.Close(); err != nil {
			logger.Warn("Error al cerrar reader de audio",
				zap.Error(err))
		}
		if err := d.vc.Speaking(false); err != nil {
			logger.Error("Error al dejar de hablar",
				zap.Error(err))
		}
	}()

	logger.Debug("Iniciando transmisión de audio")

	if err := d.vc.Speaking(true); err != nil {
		logger.Error("Error al empezar a hablar", zap.Error(err))
		return err
	}

	decoderAudio := decoder.NewBufferedOpusDecoder(reader)
	frameCount := 0

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

		select {
		case d.vc.OpusSend <- frame:
			frameCount++
		case <-ctx.Done():
			logger.Debug("Transmisión interrumpida durante envío",
				zap.String("reason", ctx.Err().Error()),
				zap.Int("frames_sent", frameCount))
			return ctx.Err()
		case <-time.After(d.sendTimeout):
			logger.Error("Timeout al enviar frame de audio",
				zap.Duration("timeout", d.sendTimeout),
				zap.Int("frames_sent", frameCount))
			return ErrSendTimeout
		}
	}
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
func (d *DiscordVoiceSession) LeaveVoiceChannel() error {
	logger := d.logger.With(
		zap.String("component", "DiscordVoiceSession"),
		zap.String("method", "LeaveVoiceChannel"),
		zap.String("guild_id", d.guildID),
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
