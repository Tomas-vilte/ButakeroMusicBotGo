package discord

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model/discord"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type DiscordMessengerAdapter struct {
	session *discordgo.Session
	logger  logging.Logger
}

func NewDiscordMessengerAdapter(session *discordgo.Session, logger logging.Logger) ports.DiscordMessenger {
	return &DiscordMessengerAdapter{
		session: session,
		logger:  logger,
	}
}

func (m *DiscordMessengerAdapter) RespondWithMessage(interaction *discord.Interaction, message string) error {
	discordInteraction := toDiscordInteraction(interaction)

	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	}
	return m.session.InteractionRespond(discordInteraction, &response)
}

func (m *DiscordMessengerAdapter) SendPlayStatus(channelID string, playMsg *entity.PlayedSong) (string, error) {
	embed := toDiscordgoEmbed(GeneratePlayingSongEmbed(playMsg))
	msg, err := m.session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		m.logger.Error("Error al enviar estado de reproducción", zap.Error(err))
		return "", err
	}
	return msg.ID, nil
}

func (m *DiscordMessengerAdapter) UpdatePlayStatus(channelID, messageID string, playMsg *entity.PlayedSong) error {
	embed := toDiscordgoEmbed(GeneratePlayingSongEmbed(playMsg))
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

func (m *DiscordMessengerAdapter) Respond(interaction *discord.Interaction, response discord.InteractionResponse) error {
	discordInteraction := toDiscordInteraction(interaction)
	discordResponse := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseType(response.Type),
		Data: &discordgo.InteractionResponseData{
			Content: response.Content,
			Embeds:  toDiscordgoEmbeds(response.Embeds),
		},
	}
	return m.session.InteractionRespond(discordInteraction, &discordResponse)
}

func (m *DiscordMessengerAdapter) CreateFollowupMessage(interaction *discord.Interaction, params discord.WebhookParams) error {
	discordInteraction := toDiscordInteraction(interaction)
	discordParams := discordgo.WebhookParams{
		Content: params.Content,
		Embeds:  toDiscordgoEmbeds(params.Embeds),
	}
	_, err := m.session.FollowupMessageCreate(discordInteraction, true, &discordParams)
	if err != nil {
		m.logger.Error("No se pudo crear el mensaje de seguimiento", zap.Error(err))
		return err
	}
	return nil
}

func (m *DiscordMessengerAdapter) EditOriginalResponse(interaction *discord.Interaction, params *discord.WebhookEdit) error {
	discordInteraction := toDiscordInteraction(interaction)
	embeds := toDiscordgoEmbeds(params.Embeds)
	discordParams := &discordgo.WebhookEdit{
		Content: params.Content,
		Embeds:  &embeds,
	}
	_, err := m.session.InteractionResponseEdit(discordInteraction, discordParams)
	if err != nil {
		m.logger.Error("No se pudo editar la respuesta original", zap.Error(err))
		return err
	}
	return nil
}

// Funciones de conversión
func toDiscordInteraction(interaction *discord.Interaction) *discordgo.Interaction {
	if interaction == nil {
		return nil
	}

	var member *discordgo.Member
	if interaction.Member != nil {
		member = &discordgo.Member{
			User: &discordgo.User{
				ID:       interaction.Member.UserID,
				Username: interaction.Member.Username,
			},
		}
	}

	return &discordgo.Interaction{
		ID:        interaction.ID,
		AppID:     interaction.AppID,
		ChannelID: interaction.ChannelID,
		GuildID:   interaction.GuildID,
		Member:    member,
		Token:     interaction.Token,
	}
}

func toDiscordgoEmbeds(embeds []*discord.Embed) []*discordgo.MessageEmbed {
	if embeds == nil {
		return nil
	}

	discordEmbeds := make([]*discordgo.MessageEmbed, len(embeds))
	for i, embed := range embeds {
		discordEmbeds[i] = toDiscordgoEmbed(embed)
	}
	return discordEmbeds
}

func toDiscordgoEmbed(embed *discord.Embed) *discordgo.MessageEmbed {
	if embed == nil {
		return nil
	}

	fields := make([]*discordgo.MessageEmbedField, len(embed.Fields))
	for i, field := range embed.Fields {
		fields[i] = &discordgo.MessageEmbedField{
			Name:   field.Name,
			Value:  field.Value,
			Inline: field.Inline,
		}
	}

	var thumbnail *discordgo.MessageEmbedThumbnail
	if embed.Thumbnail != nil {
		thumbnail = &discordgo.MessageEmbedThumbnail{
			URL:    embed.Thumbnail.URL,
			Width:  embed.Thumbnail.Width,
			Height: embed.Thumbnail.Height,
		}
	}

	var footer *discordgo.MessageEmbedFooter
	if embed.Footer != nil {
		footer = &discordgo.MessageEmbedFooter{
			Text: embed.Footer.Text,
		}
	}

	return &discordgo.MessageEmbed{
		Title:       embed.Title,
		Description: embed.Description,
		Color:       embed.Color,
		Fields:      fields,
		Thumbnail:   thumbnail,
		Footer:      footer,
	}
}
