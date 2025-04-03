package voice

import (
	"context"
	"errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/decoder"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"io"
	"time"
)

type DiscordVoiceSession struct {
	session  *discordgo.Session
	guildID  string
	vc       *discordgo.VoiceConnection
	logger   logging.Logger
	pauseCh  chan struct{}
	resumeCh chan struct{}
}

func NewDiscordVoiceSession(s *discordgo.Session, guildID string, logger logging.Logger) *DiscordVoiceSession {
	return &DiscordVoiceSession{
		session:  s,
		guildID:  guildID,
		logger:   logger,
		pauseCh:  make(chan struct{}),
		resumeCh: make(chan struct{}),
	}
}

// JoinVoiceChannel une la sesión a un canal de voz especificado por channelID.
func (d *DiscordVoiceSession) JoinVoiceChannel(channelID string) error {
	d.logger.Debug("Uniéndose al canal de voz ...", zap.String("channelID", channelID))
	vc, err := d.session.ChannelVoiceJoin(d.guildID, channelID, false, true)
	if err != nil {
		d.logger.Error("Error al unirse al canal de voz", zap.Error(err))
		return err
	}
	d.vc = vc
	return nil
}

// SendAudio envía frames de audio a la conexión de voz de Discord.
func (d *DiscordVoiceSession) SendAudio(ctx context.Context, reader io.ReadCloser) error {
	if d.vc == nil {
		return errors.New("no hay conexión de voz activa")
	}

	defer func() {
		if err := reader.Close(); err != nil {
			d.logger.Error("Error al cerrar el reader", zap.Error(err))
		}
	}()

	if err := d.vc.Speaking(true); err != nil {
		d.logger.Error("Error al comenzar a hablar", zap.Error(err))
		return err
	}
	defer func() {
		if err := d.vc.Speaking(false); err != nil {
			d.logger.Error("Error al detener la conversación", zap.Error(err))
		}
	}()

	decoderAudio := decoder.NewBufferedOpusDecoder(reader)
	isPaused := false

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-d.pauseCh:
			if !isPaused {
				isPaused = true
				d.logger.Debug("Reproducción pausada en VoiceSession")
			}
		case <-d.resumeCh:
			if isPaused {
				isPaused = false
				d.logger.Debug("Reproducción reanudada en VoiceSession")
			}
		default:
			if isPaused {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			frame, err := decoderAudio.OpusFrame()
			if err != nil {
				if err == io.EOF {
					return nil
				}
				d.logger.Error("Error al decodificar frame", zap.Error(err))
				return err
			}

			select {
			case d.vc.OpusSend <- frame:
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
				return errors.New("timeout al enviar frame de audio")
			}
		}
	}
}

func (d *DiscordVoiceSession) Pause() {
	d.pauseCh <- struct{}{}
}

func (d *DiscordVoiceSession) Resume() {
	d.resumeCh <- struct{}{}
}

// LeaveVoiceChannel desconecta la sesión del canal de voz.
func (d *DiscordVoiceSession) LeaveVoiceChannel() error {
	if d.vc != nil {
		err := d.vc.Disconnect()
		d.vc = nil
		return err
	}
	return nil
}
