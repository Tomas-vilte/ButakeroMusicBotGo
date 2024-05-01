package discord

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/codec"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"github.com/bwmarrin/discordgo"
	"io"
	"log"
	"time"
)

// VoiceChatSession representa una sesión de chat de voz en Discord.
type VoiceChatSession struct {
	discordSession  *discordgo.Session         // Sesión de Discord para enviar mensajes de texto y manejar la voz.
	guildID         string                     // ID del servidor de Discord al que pertenece la sesión.
	voiceConnection *discordgo.VoiceConnection // Conexión de voz en Discord.
}

// Close cierra la sesión de Discord.
func (session *VoiceChatSession) Close() error {
	log.Println("Cerrando sesión de Discord...")
	return session.discordSession.Close()
}

// SendMessage envía un mensaje de texto a un canal específico en Discord.
func (session *VoiceChatSession) SendMessage(channelID, message string) error {
	log.Printf("Enviando mensaje '%s' al canal %s...\n", message, channelID)
	_, err := session.discordSession.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: message,
	})
	if err != nil {
		log.Printf("Error al enviar mensaje: %v\n", err)
		return err
	}
	return nil
}

// SendPlayMessage envía un mensaje de reproducción con detalles sobre la canción que se está reproduciendo en el canal de Discord.
func (session *VoiceChatSession) SendPlayMessage(channelID string, message *bot.PlayMessage) (string, error) {
	log.Println("Enviando mensaje de reproducción...")
	// Enviar el mensaje de reproducción al canal especificado.
	msg, err := session.discordSession.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embed: GeneratePlayingSongEmbed(message),
	})
	if err != nil {
		log.Printf("Error al enviar mensaje de reproducción: %v\n", err)
		return "", err
	}
	return msg.ID, nil
}

// EditPlayMessage edita un mensaje de reproducción previamente enviado para actualizar los detalles sobre la canción que se está reproduciendo.
func (session *VoiceChatSession) EditPlayMessage(channelID string, messageID string, message *bot.PlayMessage) error {
	// Editar el mensaje de reproducción con los nuevos detalles de la canción.
	embeds := []*discordgo.MessageEmbed{GeneratePlayingSongEmbed(message)}
	_, err := session.discordSession.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:      messageID,
		Channel: channelID,
		Embeds:  &embeds,
	})
	if err != nil {
		log.Printf("Error al editar mensaje de reproducción: %v\n", err)
	}

	return err
}

// JoinVoiceChannel se une a un canal de voz en Discord.
func (session *VoiceChatSession) JoinVoiceChannel(channelID string) error {
	log.Printf("Uniéndose al canal de voz %s...\n", channelID)
	// Unirse al canal de voz en Discord.
	vc, err := session.discordSession.ChannelVoiceJoin(session.guildID, channelID, false, true)
	if err != nil {
		log.Printf("Error al unirse al canal de voz: %v\n", err)
		return fmt.Errorf("mientras se unía al canal de voz: %w", err)
	}
	session.voiceConnection = vc
	return nil
}

// LeaveVoiceChannel abandona el canal de voz en Discord.
func (session *VoiceChatSession) LeaveVoiceChannel() error {
	log.Println("Dejando el canal de voz...")
	if session.voiceConnection == nil {
		return nil
	}

	// Dejar el canal de voz en Discord.
	if err := session.voiceConnection.Disconnect(); err != nil {
		log.Printf("Error al dejar el canal de voz: %v\n", err)
		return err
	}

	session.voiceConnection = nil
	return nil
}

// SendAudio envía datos de audio a través de la conexión de voz en Discord utilizando el códec DCA.
func (session *VoiceChatSession) SendAudio(ctx context.Context, reader io.Reader, positionCallback func(time.Duration)) error {
	log.Println("Enviando audio al canal de voz...")
	// Indicar que el bot está hablando en el canal de voz.
	if err := session.voiceConnection.Speaking(true); err != nil {
		log.Printf("Error al comenzar a hablar: %v\n", err)
		return fmt.Errorf("mientras se comenzaba a hablar: %w", err)
	}

	// Transmitir los datos de audio utilizando el códec DCA.
	if err := codec.StreamDCAData(ctx, reader, session.voiceConnection.OpusSend, positionCallback); err != nil {
		log.Printf("Error al transmitir datos DCA: %v\n", err)
		return fmt.Errorf("mientras se transmitían datos DCA: %w", err)
	}

	// Indicar que el bot ha dejado de hablar en el canal de voz.
	if err := session.voiceConnection.Speaking(false); err != nil {
		log.Printf("Error al dejar de hablar: %v\n", err)
		return fmt.Errorf("mientras se dejaba de hablar: %w", err)
	}
	return nil
}
