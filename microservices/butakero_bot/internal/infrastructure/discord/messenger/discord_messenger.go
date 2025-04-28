package messenger

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/infrastructure/discord/interfaces"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type DiscordMessengerAdapter struct {
	session *discordgo.Session
	logger  logging.Logger
}

func NewDiscordMessengerAdapter(session *discordgo.Session, logger logging.Logger) interfaces.DiscordMessenger {
	return &DiscordMessengerAdapter{
		session: session,
		logger:  logger,
	}
}

func (m *DiscordMessengerAdapter) RespondWithMessage(interaction *discordgo.Interaction, message string) error {
	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	}
	return m.session.InteractionRespond(interaction, &response)
}

func (m *DiscordMessengerAdapter) SendPlayStatus(channelID string, playMsg *entity.PlayedSong) (string, error) {
	embed := discord.GeneratePlayingSongEmbed(playMsg)
	msg, err := m.session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		m.logger.Error("Error al enviar estado de reproducción", zap.Error(err))
		return "", err
	}
	return msg.ID, nil
}

func (m *DiscordMessengerAdapter) UpdatePlayStatus(channelID, messageID string, playMsg *entity.PlayedSong) error {
	embed := discord.GeneratePlayingSongEmbed(playMsg)
	_, err := m.session.ChannelMessageEditEmbed(channelID, messageID, embed)
	if err != nil {
		m.logger.Error("Error al actualizar estado de reproducción", zap.Error(err))
	}
	return err
}

func (m *DiscordMessengerAdapter) SendText(channelID, text string) error {
	_, err := m.session.ChannelMessageSend(channelID, text)
	if err != nil {
		m.logger.Error("Error al enviar mensaje de texto", zap.Error(err))
	}
	return err
}

func (m *DiscordMessengerAdapter) Respond(interaction *discordgo.Interaction, response *discordgo.InteractionResponse) error {
	return m.session.InteractionRespond(interaction, response)
}

func (m *DiscordMessengerAdapter) CreateFollowupMessage(interaction *discordgo.Interaction, params *discordgo.WebhookParams) error {
	_, err := m.session.FollowupMessageCreate(interaction, true, params)
	if err != nil {
		m.logger.Error("No se pudo crear el mensaje de seguimiento", zap.Error(err))
		return err
	}
	return nil
}

func (m *DiscordMessengerAdapter) EditOriginalResponse(interaction *discordgo.Interaction, params *discordgo.WebhookEdit) error {
	_, err := m.session.InteractionResponseEdit(interaction, params)
	if err != nil {
		m.logger.Error("No se pudo editar la respuesta original", zap.Error(err))
		return err
	}
	return nil
}
