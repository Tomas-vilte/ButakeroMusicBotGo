package voice

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type DiscordVoiceSession struct {
	session *discordgo.Session
	guildID string
	vc      *discordgo.VoiceConnection
	logger  logging.Logger
}

func NewDiscordVoiceSession(s *discordgo.Session, guildID string, logger logging.Logger) *DiscordVoiceSession {
	return &DiscordVoiceSession{
		session: s,
		guildID: guildID,
		logger:  logger,
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
func (d *DiscordVoiceSession) SendAudio(ctx context.Context, frames []byte) error {
	if err := d.vc.Speaking(true); err != nil {
		d.logger.Error("Error al comenzar a hablar: ", zap.Error(err))
		return err
	}

	defer func() {
		if err := d.vc.Speaking(false); err != nil {
			d.logger.Error("Error al dejar de hablar: ", zap.Error(err))
			return
		}
	}()

	select {
	case d.vc.OpusSend <- frames:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close cierra la conexión de voz de Discord.
func (d *DiscordVoiceSession) Close() error {
	if d.vc != nil {
		return d.vc.Disconnect()
	}
	return nil
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
