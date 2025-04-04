package voice

import (
	"context"
	"errors"
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
		sendTimeout: 1 * time.Second,
	}
}

// JoinVoiceChannel conecta la sesión a un canal de voz específico usando channelID
func (d *DiscordVoiceSession) JoinVoiceChannel(channelID string) error {
	d.logger.Debug("Conectando al canal de voz", zap.String("channelID", channelID))
	vc, err := d.session.ChannelVoiceJoin(d.guildID, channelID, false, true)
	if err != nil {
		d.logger.Error("Falló la conexión al canal de voz", zap.Error(err))
		return err
	}
	d.vc = vc
	return nil
}

// SendAudio manda frames de audio a la conexión de voz de Discord
func (d *DiscordVoiceSession) SendAudio(ctx context.Context, reader io.ReadCloser) error {
	if d.vc == nil {
		return ErrNoVoiceConnection
	}

	defer func() {
		_ = reader.Close()
		if err := d.vc.Speaking(false); err != nil {
			d.logger.Error("Error al dejar de hablar", zap.Error(err))
		}
	}()

	if err := d.vc.Speaking(true); err != nil {
		d.logger.Error("Error al empezar a hablar", zap.Error(err))
		return err
	}

	decoderAudio := decoder.NewBufferedOpusDecoder(reader)

	for {
		select {
		case <-ctx.Done():
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
				return nil
			}
			d.logger.Error("Error decodificando el frame", zap.Error(err))
			return err
		}

		select {
		case d.vc.OpusSend <- frame:
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(d.sendTimeout):
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
	if d.vc != nil {
		d.logger.Debug("Saliendo del canal de voz")
		err := d.vc.Disconnect()
		d.vc = nil
		return err
	}
	return nil
}

// SetSendTimeout configura el timeout para enviar frames de audio
func (d *DiscordVoiceSession) SetSendTimeout(timeout time.Duration) {
	d.sendTimeout = timeout
}
