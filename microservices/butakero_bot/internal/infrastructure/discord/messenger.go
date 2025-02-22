package discord

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

// MessengerService implementa la interfaz DiscordMessenger.
type MessengerService struct {
	session *discordgo.Session
	logger  logging.Logger
}

func NewDiscordMessengerService(session *discordgo.Session, logger logging.Logger) ports.DiscordMessenger {
	return &MessengerService{
		session: session,
		logger:  logger,
	}
}

// RespondToInteraction Responde a una interacción con un embed.
func (m *MessengerService) RespondToInteraction(interaction *discordgo.Interaction, embed *discordgo.MessageEmbed) error {
	m.logger.Info("Respondiendo a interacción", zap.String("tipo", interaction.Type.String()))

	err := m.session.InteractionRespond(interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})

	if err != nil {
		m.logger.Error("Error al responder a la interacción", zap.Error(err))
	}
	return err
}

// SendPlayStatus Envía un embed de estado de reproducción a un canal.
func (m *MessengerService) SendPlayStatus(channelID string, playMsg *entity.PlayedSong) (string, error) {
	m.logger.Info("Enviando estado de reproducción", zap.String("channelID", channelID))

	embed := GeneratePlayingSongEmbed(playMsg)
	msg, err := m.session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		m.logger.Error("Error al enviar estado de reproducción", zap.Error(err))
		return "", err
	}
	return msg.ID, nil
}

// UpdatePlayStatus Actualiza un mensaje de estado existente.
func (m *MessengerService) UpdatePlayStatus(channelID, messageID string, playMsg *entity.PlayedSong) error {
	m.logger.Info("Actualizando estado de reproducción", zap.String("messageID", messageID))

	embed := GeneratePlayingSongEmbed(playMsg)
	_, err := m.session.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		m.logger.Error("Error al actualizar estado de reproducción", zap.Error(err))
	}
	return err
}

// SendText Envía un mensaje de texto simple.
func (m *MessengerService) SendText(channelID, text string) error {
	m.logger.Info("Enviando mensaje de texto", zap.String("channelID", channelID))

	_, err := m.session.ChannelMessageSend(channelID, text)
	if err != nil {
		m.logger.Error("Error al enviar mensaje de texto", zap.Error(err))
	}
	return err
}
