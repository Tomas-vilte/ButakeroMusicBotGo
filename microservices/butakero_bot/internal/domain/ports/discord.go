package ports

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/bwmarrin/discordgo"
)

// DiscordMessenger define todas las operaciones de mensajería relacionadas con Discord.
type DiscordMessenger interface {
	// RespondToInteraction Responde a una interacción (comando slash, botón, etc.) con un embed.
	RespondToInteraction(interaction *discordgo.Interaction, embed *discordgo.MessageEmbed) error
	// SendPlayStatus Envía un mensaje embed de estado de reproducción (ej: "Ahora sonando").
	SendPlayStatus(channelID string, playMsg *entity.PlayedSong) (messageID string, err error)
	// UpdatePlayStatus Actualiza un mensaje de estado de reproducción existente.
	UpdatePlayStatus(channelID, messageID string, playMsg *entity.PlayedSong) error
	// SendText Envía un mensaje de texto simple a un canal.
	SendText(channelID, text string) error
}
