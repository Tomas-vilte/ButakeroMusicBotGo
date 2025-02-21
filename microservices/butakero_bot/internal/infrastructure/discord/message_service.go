package discord

import (
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared"
	"github.com/bwmarrin/discordgo"
)

type MessageService struct {
}

func NewMessageService() *MessageService {
	return &MessageService{}
}

// GenerateAddedSongEmbed Genera un embed cuando se añade una canción a la cola.
func (s *MessageService) GenerateAddedSongEmbed(song *entity.Song, member *discordgo.Member) *discordgo.MessageEmbed {
	embed := s.generateBaseEmbed(song.Title, "🎵  Agregado a la cola.", member)
	embed.Fields = []*discordgo.MessageEmbedField{
		{Name: "Duración", Value: shared.FmtDuration(song.Duration)},
	}
	if song.Thumbnail != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{URL: song.Thumbnail}
	}
	return embed
}

// GenerateFailedToAddSongEmbed Genera un embed cuando falla la adición de una canción.
func (s *MessageService) GenerateFailedToAddSongEmbed(input string, member *discordgo.Member) *discordgo.MessageEmbed {
	return s.generateBaseEmbed(input, "❌ Error al añadir la canción", member)
}

// generateBaseEmbed Función base para generar embeds.
func (s *MessageService) generateBaseEmbed(title, description string, member *discordgo.Member) *discordgo.MessageEmbed {
	footerText := "Usuario desconocido"
	if member != nil && member.User != nil {
		footerText = member.User.Username
	}
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Pedido por: %s", footerText)},
	}
}
