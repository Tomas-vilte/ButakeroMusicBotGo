package voice

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

// MessageSenderWrapper es una interfaz que envuelve los métodos necesarios de discordgo.Session para enviar mensajes.
type MessageSenderWrapper interface {
	ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error)
	ChannelMessageEditComplex(m *discordgo.MessageEdit, options ...discordgo.RequestOption) (*discordgo.Message, error)
}

// MessageSenderWrapperImpl es una implementación concreta de MessageSenderWrapper que envuelve una instancia de discordgo.Session.
type MessageSenderWrapperImpl struct {
	session *discordgo.Session
}

func (w *MessageSenderWrapperImpl) ChannelMessageSendComplex(channelID string, data *discordgo.MessageSend, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	return w.session.ChannelMessageSendComplex(channelID, data, options...)
}

func (w *MessageSenderWrapperImpl) ChannelMessageEditComplex(m *discordgo.MessageEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	return w.session.ChannelMessageEditComplex(m, options...)
}

// ChatMessageSender envía mensajes de chat a Discord.
type ChatMessageSender interface {
	SendMessage(channelID, message string) error
	SendPlayMessage(channelID string, message *PlayMessage) (string, error)
	EditPlayMessage(channelID, messageID string, message *PlayMessage) error
}

// MessageSenderImpl implementa la interfaz ChatMessageSender para enviar mensajes en Discord.
type MessageSenderImpl struct {
	DiscordSession MessageSenderWrapper
}

// SendMessage envía un mensaje de texto a un canal específico en Discord.
func (session *MessageSenderImpl) SendMessage(channelID, message string) error {
	log.Printf("Enviando mensaje '%s' al canal %s...\n", message, channelID)
	_, err := session.DiscordSession.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: message,
	})
	if err != nil {
		log.Printf("Error al enviar mensaje: %v\n", err)
		return err
	}
	return nil
}

// SendPlayMessage envía un mensaje de reproducción con detalles sobre la canción que se está reproduciendo en el canal de Discord.
func (session *MessageSenderImpl) SendPlayMessage(channelID string, message *PlayMessage) (string, error) {
	log.Println("Enviando mensaje de reproducción...")
	// Enviar el mensaje de reproducción al canal especificado.
	msg, err := session.DiscordSession.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embed: GeneratePlayingSongEmbed(message),
	})
	if err != nil {
		log.Printf("Error al enviar mensaje de reproducción: %v\n", err)
		return "", err
	}
	return msg.ID, nil
}

// EditPlayMessage edita un mensaje de reproducción previamente enviado para actualizar los detalles sobre la canción que se está reproduciendo.
func (session *MessageSenderImpl) EditPlayMessage(channelID string, messageID string, message *PlayMessage) error {
	// Editar el mensaje de reproducción con los nuevos detalles de la canción.
	embeds := []*discordgo.MessageEmbed{GeneratePlayingSongEmbed(message)}
	_, err := session.DiscordSession.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:      messageID,
		Channel: channelID,
		Embeds:  &embeds,
	})
	if err != nil {
		log.Printf("Error al editar mensaje de reproducción: %v\n", err)
	}

	return err
}
