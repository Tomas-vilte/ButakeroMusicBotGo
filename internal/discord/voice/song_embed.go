package voice

import (
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/utils"
	"github.com/bwmarrin/discordgo"
)

// GeneratePlayingSongEmbed un mensaje embed para mostrar que se está agregando una canción a la cola de reproducción.
func GeneratePlayingSongEmbed(message *PlayMessage) *discordgo.MessageEmbed {
	if message == nil || message.Song == nil {
		return nil // Retornamos nil si message o message.Song es nil
	}

	progressBar := generateProgressBar(float64(message.Position)/float64(message.Song.Duration), 20)

	embed := &discordgo.MessageEmbed{
		Title:       message.Song.GetHumanName(),
		Description: fmt.Sprintf("%s\n%s / %s", progressBar, utils.FmtDuration(message.Position), utils.FmtDuration(message.Song.Duration)),
	}
	if message.Song.ThumbnailURL != nil {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: *message.Song.ThumbnailURL,
		}
	}

	if message.Song.RequestedBy != nil {
		embed.Footer = &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Solicitado por: %v", *message.Song.RequestedBy),
		}
	}
	return embed
}

func generateProgressBar(progress float64, length int) string {
	played := int(progress * float64(length))

	progressBar := ""
	for i := 0; i < played; i++ {
		progressBar += "▬"
	}
	progressBar += "🔘"
	for i := 0; i < length; i++ {
		progressBar += "▬"
	}
	return progressBar
}
