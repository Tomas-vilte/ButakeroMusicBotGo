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

// RespondWithMessage responde a una interacción de Discord con un mensaje de texto.
func (m *MessengerService) RespondWithMessage(interaction *discordgo.Interaction, message string) error {
	m.logger.Info("Respondiendo a interacción", zap.String("tipo", interaction.Type.String()))

	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	}
	return m.Respond(interaction, response)
}

// Respond responde a una interacción de Discord con la respuesta proporcionada.
func (m *MessengerService) Respond(interaction *discordgo.Interaction, response discordgo.InteractionResponse) error {
	if err := m.session.InteractionRespond(interaction, &response); err != nil {
		m.logger.Error("No se pudo responder a la interacción", zap.Error(err))
		return err
	}
	return nil
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

// CreateFollowupMessage crea un mensaje de seguimiento para una interacción de Discord.
func (m *MessengerService) CreateFollowupMessage(interaction *discordgo.Interaction, params discordgo.WebhookParams) error {
	if _, err := m.session.FollowupMessageCreate(interaction, true, &params); err != nil {
		m.logger.Error("No se pudo crear el mensaje de seguimiento", zap.Error(err))
		return err
	}
	return nil
}

func (m *MessengerService) EditOriginalResponse(interaction *discordgo.Interaction, params *discordgo.WebhookEdit) error {
	_, err := m.session.InteractionResponseEdit(interaction, params)
	if err != nil {
		m.logger.Error("No se pudo editar la respuesta original", zap.Error(err))
		return err
	}
	return nil
}
