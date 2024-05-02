package discord

import (
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"github.com/Tomas-vilte/GoMusicBot/internal/utils"
	"github.com/bwmarrin/discordgo"
)

const (
	ErrorMessageNotInVoiceChannel = "No estas en un canal de voz down. Tenes que unirte a uno para reproducir musica loco"
	ErrorMessageFailedToAddSong   = "No se pudo agregar la cancion kkkk"
)

func GenerateAddingSongEmbed(input string, member *discordgo.Member) *discordgo.MessageEmbed {
	return generateAddingSongEmbed(input, "🎵  Añadiendo cancion a la cola...", member)
}

func GenerateFailedToAddSongEmbed(input string, member *discordgo.Member) *discordgo.MessageEmbed {
	return generateAddingSongEmbed(input, "😨 Error al añadir la cancion a la cola", member)
}

func GenerateFailedToFindSong(input string, member *discordgo.Member) *discordgo.MessageEmbed {
	return generateAddingSongEmbed(input, "😨 No se pudo encontrar ninguna canción reproducible.", member)
}

func GenerateAskAddPlaylistEmbed(songs []*bot.Song, requestor *discordgo.Member) *discordgo.MessageEmbed {
	title := fmt.Sprintf("👀  La canción es parte de una lista de reproducción que contiene %d canciones. Que mierda hago?", len(songs))
	return generateAddingSongEmbed(title, "", requestor)
}

func GenerateAddedSongEmbed(song *bot.Song, member *discordgo.Member) *discordgo.MessageEmbed {
	embed := generateAddingSongEmbed(song.GetHumanName(), "🎵  Agregado a la cola.", member)
	embed.Fields = []*discordgo.MessageEmbedField{
		{
			Name:  "Duracion",
			Value: utils.FmtDuration(song.Duration),
		},
	}

	if song.ThumbnailURL != nil {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: *song.ThumbnailURL,
		}
	}

	return embed
}

// GeneratePlayingSongEmbed un mensaje embed para mostrar que se está agregando una canción a la cola de reproducción.
func GeneratePlayingSongEmbed(message *bot.PlayMessage) *discordgo.MessageEmbed {
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

func generateAddingSongEmbed(title, description string, requestor *discordgo.Member) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Pedido por: %s", getMemberName(requestor)),
		},
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
