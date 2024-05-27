package discordmessenger

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
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
	SendPlayMessage(channelID string, message *voice.PlayMessage) (string, error)
	EditPlayMessage(channelID, messageID string, message *voice.PlayMessage) error
}

// MessageSenderImpl implementa la interfaz ChatMessageSender para enviar mensajes en Discord.
type MessageSenderImpl struct {
	DiscordSession MessageSenderWrapper
	logger         logging.Logger
}

func NewMessageSenderImpl(discordSession MessageSenderWrapper, logger logging.Logger) *MessageSenderImpl {
	return &MessageSenderImpl{
		DiscordSession: discordSession,
		logger:         logger,
	}
}

// SendMessage envía un mensaje de texto a un canal específico en Discord.
func (session *MessageSenderImpl) SendMessage(channelID, message string) error {
	session.logger.Info("Enviando mensaje al canal", zap.String("mensaje", message), zap.String("channel", channelID))
	_, err := session.DiscordSession.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: message,
	})
	if err != nil {
		session.logger.Error("Error al enviar el mensaje: ", zap.Error(err))
		return err
	}
	return nil
}

// SendPlayMessage envía un mensaje de reproducción con detalles sobre la canción que se está reproduciendo en el canal de Discord.
func (session *MessageSenderImpl) SendPlayMessage(channelID string, message *voice.PlayMessage) (string, error) {
	session.logger.Info("Enviando mensaje de reproducción...")
	// Enviar el mensaje de reproducción al canal especificado.
	msg, err := session.DiscordSession.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Embed: voice.GeneratePlayingSongEmbed(message),
	})
	if err != nil {
		session.logger.Error("Error al enviar mensaje de reproducción: ", zap.Error(err))
		return "", err
	}
	return msg.ID, nil
}

// EditPlayMessage edita un mensaje de reproducción previamente enviado para actualizar los detalles sobre la canción que se está reproduciendo.
func (session *MessageSenderImpl) EditPlayMessage(channelID string, messageID string, message *voice.PlayMessage) error {
	// Editar el mensaje de reproducción con los nuevos detalles de la canción.
	embeds := []*discordgo.MessageEmbed{voice.GeneratePlayingSongEmbed(message)}
	_, err := session.DiscordSession.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:      messageID,
		Channel: channelID,
		Embeds:  &embeds,
	})
	if err != nil {
		session.logger.Error("Error al editar el mensaje de reproducción: ", zap.Error(err))
		return err
	}

	return err
}
