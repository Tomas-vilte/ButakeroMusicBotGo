package ports

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/bwmarrin/discordgo"
)

// DiscordResponseHandler define cómo responder a interacciones de Discord.
type DiscordResponseHandler interface {
	Respond(interaction *discordgo.Interaction, response discordgo.InteractionResponse) error
	RespondWithMessage(interaction *discordgo.Interaction, message string) error
	CreateFollowupMessage(interaction *discordgo.Interaction, params discordgo.WebhookParams) error
}

// DiscordMessageService define cómo generar mensajes embebidos para Discord.
type DiscordMessageService interface {
	GenerateAddedSongEmbed(song *entity.Song, member *discordgo.Member) *discordgo.MessageEmbed
	GenerateFailedToAddSongEmbed(input string, member *discordgo.Member) *discordgo.MessageEmbed
}
