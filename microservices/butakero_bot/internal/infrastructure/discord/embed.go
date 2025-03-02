package discord

import (
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared"
	"github.com/bwmarrin/discordgo"
	"time"
)

// GeneratePlayingSongEmbed Genera un embed para mostrar una canción en reproducción.
func GeneratePlayingSongEmbed(playMsg *entity.PlayedSong) *discordgo.MessageEmbed {
	if playMsg == nil || playMsg.Song == (entity.Song{}) {
		return nil
	}

	progressBar := generateProgressBar(
		float64(playMsg.Position)/parseDuration(playMsg.Song.Duration),
		20,
	)

	embed := &discordgo.MessageEmbed{
		Title:       playMsg.Song.Title,
		Description: fmt.Sprintf("%s\n%s / %s", progressBar, shared.FmtDuration(playMsg.Position), shared.FmtDuration(time.Duration(parseDuration(playMsg.Song.Duration))))}

	if playMsg.Song.ThumbnailURL != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: playMsg.Song.ThumbnailURL,
		}
	}

	if playMsg.RequestedBy != "" {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Solicitado por: %s", playMsg.RequestedBy),
		}
	}

	return embed
}

func parseDuration(duration string) float64 {
	d, _ := time.ParseDuration(duration)
	return d.Seconds()
}

// Función interna para generar la barra de progreso.
func generateProgressBar(progress float64, length int) string {
	filled := int(progress * float64(length))
	bar := ""
	for i := 0; i < filled; i++ {
		bar += "🟥"
	}
	bar += "🔴"
	for i := 0; i < length-filled-1; i++ {
		bar += "⬛"
	}
	return bar
}
