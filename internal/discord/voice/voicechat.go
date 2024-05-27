package voice

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice/codec"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
	"io"
	"time"
)

// DiscordSessionWrapper es una interfaz que envuelve los métodos de discordgo.Session que necesitamos mockear.
type DiscordSessionWrapper interface {
	ChannelVoiceJoin(guildID, channelID string, muted, deafened bool) (*discordgo.VoiceConnection, error)
	Close() error
}

// ConnectionWrapper es una interfaz que envuelve los métodos de discordgo.VoiceConnection que necesitamos mockear.
type ConnectionWrapper interface {
	Disconnect() error
	Speaking(flag bool) error
	OpusSend(data []byte, mode int) (ok bool, err error)
	OpusSendChan() chan<- []byte
}

// DiscordSessionWrapperImpl es una implementación concreta de DiscordSessionWrapper que envuelve una instancia de discordgo.Session.
type DiscordSessionWrapperImpl struct {
	session *discordgo.Session
}

func (w *DiscordSessionWrapperImpl) ChannelVoiceJoin(guildID, channelID string, muted, deafened bool) (*discordgo.VoiceConnection, error) {
	return w.session.ChannelVoiceJoin(guildID, channelID, muted, deafened)
}

func (w *DiscordSessionWrapperImpl) Close() error {
	return w.session.Close()
}

// ConnectionWrapperImpl es una implementación concreta de ConnectionWrapper que envuelve una instancia de discordgo.VoiceConnection.
type ConnectionWrapperImpl struct {
	voiceConnection *discordgo.VoiceConnection
	opusSendChan    chan []byte
}

func (w *ConnectionWrapperImpl) Disconnect() error {
	if w.voiceConnection == nil {
		return nil
	}
	return w.voiceConnection.Disconnect()
}

func (w *ConnectionWrapperImpl) Speaking(flag bool) error {
	return w.voiceConnection.Speaking(flag)
}

func (w *ConnectionWrapperImpl) OpusSend(data []byte, mode int) (bool, error) {
	w.opusSendChan <- data
	return true, nil
}

func (w *ConnectionWrapperImpl) OpusSendChan() chan<- []byte {
	w.opusSendChan = w.voiceConnection.OpusSend
	return w.opusSendChan
}

// ChatSessionImpl representa una sesión de chat de voz en Discord.
type ChatSessionImpl struct {
	DiscordSession  DiscordSessionWrapper // Sesión de Discord para enviar mensajes de texto y manejar la voz.
	GuildID         string                // ID del servidor de Discord al que pertenece la sesión.
	voiceConnection ConnectionWrapper     // Conexión de voz en Discord.
	DCAStreamer     codec.DCAStreamer
	logger          logging.Logger
}

func NewChatSessionImpl(discordSessionWrapper DiscordSessionWrapper, guildID string, DCAStreamer codec.DCAStreamer, logger logging.Logger) *ChatSessionImpl {
	return &ChatSessionImpl{
		DiscordSession: discordSessionWrapper,
		GuildID:        guildID,
		DCAStreamer:    DCAStreamer,
		logger:         logger,
	}
}

// Close cierra la sesión de Discord.
func (session *ChatSessionImpl) Close() error {
	session.logger.Info("Cerrando sesión de Discord...")
	return session.DiscordSession.Close()
}

// JoinVoiceChannel se une a un canal de voz en Discord.
func (session *ChatSessionImpl) JoinVoiceChannel(channelID string) error {
	session.logger.Info("Uniéndose al canal de voz ...", zap.String("channelID", channelID))
	// Unirse al canal de voz en Discord.
	vc, err := session.DiscordSession.ChannelVoiceJoin(session.GuildID, channelID, false, true)
	if err != nil {
		session.logger.Error("Error al unirse al canal de voz", zap.Error(err))
		return err
	}
	session.voiceConnection = &ConnectionWrapperImpl{
		voiceConnection: vc,
	}
	return nil
}

// LeaveVoiceChannel abandona el canal de voz en Discord.
func (session *ChatSessionImpl) LeaveVoiceChannel() error {
	if session.voiceConnection == nil {
		return nil
	}

	// Dejar el canal de voz en Discord.
	err := session.voiceConnection.Disconnect()
	session.voiceConnection = nil

	if err != nil {
		session.logger.Error("Error al dejar el canal de voz", zap.Error(err))
		return err
	}

	return nil
}

// SendAudio envía datos de audio a través de la conexión de voz en Discord utilizando el códec DCA.
func (session *ChatSessionImpl) SendAudio(ctx context.Context, reader io.Reader, positionCallback func(time.Duration)) error {
	session.logger.Info("Enviando audio al canal de voz...")

	if err := session.voiceConnection.Speaking(true); err != nil {
		session.logger.Error("Error al comenzar a hablar: ", zap.Error(err))
		return err
	}

	opusSendChan := session.voiceConnection.OpusSendChan()
	if opusSendChan == nil {
		session.logger.Error("Error canal de envío de Opus no está disponible")
		return fmt.Errorf("canal de envío de Opus no está disponible")
	}

	if err := session.DCAStreamer.StreamDCAData(ctx, reader, opusSendChan, positionCallback); err != nil {
		session.logger.Error("Error al transmitir datos DCA: ", zap.Error(err))
		_ = session.voiceConnection.Speaking(false)
		return err
	}

	if err := session.voiceConnection.Speaking(false); err != nil {
		session.logger.Error("Error al dejar de hablar: ", zap.Error(err))
		return err
	}

	return nil
}
