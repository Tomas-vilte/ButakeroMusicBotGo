package interfaces

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/bwmarrin/discordgo"
)

type DiscordMessenger interface {
	// RespondWithMessage responde a una interacción con un mensaje
	RespondWithMessage(interaction *discordgo.Interaction, message string) error
	// SendPlayStatus envía un mensaje embed de estado de reproducción
	SendPlayStatus(channelID string, playMsg *entity.PlayedSong) (messageID string, err error)
	// UpdatePlayStatus actualiza un mensaje de estado existente
	UpdatePlayStatus(channelID, messageID string, playMsg *entity.PlayedSong) error
	// SendText envía un mensaje de texto simple
	SendText(channelID, text string) error
	// Respond responde a una interacción con una respuesta estructurada
	Respond(interaction *discordgo.Interaction, response *discordgo.InteractionResponse) error
	// CreateFollowupMessage crea un mensaje de seguimiento
	CreateFollowupMessage(interaction *discordgo.Interaction, params *discordgo.WebhookParams) error
	// EditOriginalResponse edita la respuesta original
	EditOriginalResponse(interaction *discordgo.Interaction, params *discordgo.WebhookEdit) error
}
