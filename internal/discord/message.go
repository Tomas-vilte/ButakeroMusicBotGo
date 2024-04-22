package discord

import (
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"github.com/Tomas-vilte/GoMusicBot/internal/utils"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

const (
	ErrorMessageNotInVoiceChannel = "No estas en un canal de voz down. Tenes que unirte a uno para reproducir musica loco"
	ErrorMessageFailedToAddSong   = "No se pudo agregar la cancion kkkk"
	ErrorMessageFailedToFindSong  = "No se encontraron canciones reproducibles kkkk"
)

const (
	EmbedMessageAddingSong  = "Agregando musica a la cola"
	EmbedMessageAddedSong   = "Agregada en cola"
	EmbedMessageFailedToAdd = "No se pudo agregar la cola"
)

// GeneratePlayingSongEmbed un mensaje embed para mostrar que se está agregando una canción a la cola de reproducción.
func GeneratePlayingSongEmbed(message *bot.PlayMessage) *discordgo.MessageEmbed {
	// TODO: impl
	progressBar := generateProgressBar(float64(message.Position)/float64(message.Song.Duration), 20)

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s", message.Song.GetHumanName()),
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

func GeneratePlaylistAdded(intro string, songs []*bot.Song, member *discordgo.Member) *discordgo.MessageEmbed {
	descriptionBuilder := strings.Builder{}
	duration := time.Duration(0)

	for _, song := range songs {
		duration += song.Duration
		descriptionBuilder.WriteString(fmt.Sprintf("1. %s (%s)\n", song.GetHumanName(), utils.FmtDuration(song.Duration)))
	}

	title := fmt.Sprintf("%s", intro)
	embed := generateAddingSongEmbed(title, descriptionBuilder.String(), member)
	embed.Fields = []*discordgo.MessageEmbedField{
		{Name: "Duracion"},
		{Value: utils.FmtDuration(duration)},
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
