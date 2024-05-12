package voice

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice/codec"
	"github.com/bwmarrin/discordgo"
	"io"
	"log"
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
}

// Close cierra la sesión de Discord.
func (session *ChatSessionImpl) Close() error {
	log.Println("Cerrando sesión de Discord...")
	return session.DiscordSession.Close()
}

// JoinVoiceChannel se une a un canal de voz en Discord.
func (session *ChatSessionImpl) JoinVoiceChannel(channelID string) error {
	log.Printf("Uniéndose al canal de voz %s...\n", channelID)
	// Unirse al canal de voz en Discord.
	vc, err := session.DiscordSession.ChannelVoiceJoin(session.GuildID, channelID, false, true)
	if err != nil {
		log.Printf("Error al unirse al canal de voz: %v\n", err)
		return fmt.Errorf("mientras se unía al canal de voz: %w", err)
	}
	session.voiceConnection = &ConnectionWrapperImpl{
		voiceConnection: vc,
	}
	return nil
}

// LeaveVoiceChannel abandona el canal de voz en Discord.
func (session *ChatSessionImpl) LeaveVoiceChannel() error {
	log.Println("Dejando el canal de voz...")
	if session.voiceConnection == nil {
		return nil
	}

	// Dejar el canal de voz en Discord.
	err := session.voiceConnection.Disconnect()
	session.voiceConnection = nil

	if err != nil {
		log.Printf("Error al dejar el canal de voz: %v\n", err)
		return err
	}

	return nil
}

// SendAudio envía datos de audio a través de la conexión de voz en Discord utilizando el códec DCA.
func (session *ChatSessionImpl) SendAudio(ctx context.Context, reader io.Reader, positionCallback func(time.Duration)) error {
	log.Println("Enviando audio al canal de voz...")

	if err := session.voiceConnection.Speaking(true); err != nil {
		log.Printf("Error al comenzar a hablar: %v\n", err)
		return fmt.Errorf("mientras se comenzaba a hablar: %w", err)
	}

	if err := session.DCAStreamer.StreamDCAData(ctx, reader, session.voiceConnection.OpusSendChan(), positionCallback); err != nil {
		log.Printf("Error al transmitir datos DCA: %v\n", err)
		session.voiceConnection.Speaking(false)
		return fmt.Errorf("mientras se transmitían datos DCA: %w", err)
	}

	if err := session.voiceConnection.Speaking(false); err != nil {
		log.Printf("Error al dejar de hablar: %v\n", err)
		return fmt.Errorf("mientras se dejaba de hablar: %w", err)
	}
	return nil
}
