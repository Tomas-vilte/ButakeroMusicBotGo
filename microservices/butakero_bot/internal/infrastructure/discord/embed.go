package discord

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/bwmarrin/discordgo"
)

// GeneratePlayingSongEmbed Genera un embed para mostrar una canción en reproducción.
func GeneratePlayingSongEmbed(playMsg *entity.PlayedSong) *discordgo.MessageEmbed {
	//if playMsg == nil || playMsg.Song == (entity.Song{}) {
	//	return nil
	//}
	//
	//progressBar := generateProgressBar(
	//	float64(playMsg.Position)/float64(playMsg.Song.Duration),
	//	20,
	//)
	//
	//embed := &discordgo.MessageEmbed{
	//	Title:       playMsg.Song.Title,
	//	Description: fmt.Sprintf("%s\n%s / %s", progressBar, shared.FmtDuration(playMsg.Position), shared.FmtDuration(playMsg.Song.Duration)),
	//}
	//
	//if playMsg.Song.ThumbnailURL != "" {
	//	embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
	//		URL: playMsg.Song.ThumbnailURL,
	//	}
	//}
	//
	//if playMsg.RequestedBy != "" {
	//	embed.Footer = &discordgo.MessageEmbedFooter{
	//		Text: fmt.Sprintf("Solicitado por: %s", playMsg.RequestedBy),
	//	}
	//}
	//
	//return embed
	return nil
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
