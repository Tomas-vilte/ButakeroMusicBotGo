package discord

import (
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
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

func GenerateAskAddPlaylistEmbed(songs []*voice.Song, requestor *discordgo.Member) *discordgo.MessageEmbed {
	title := fmt.Sprintf("👀  La canción es parte de una lista de reproducción que contiene %d canciones. Que mierda hago?", len(songs))
	return generateAddingSongEmbed(title, "", requestor)
}

func GenerateAddedSongEmbed(song *voice.Song, member *discordgo.Member) *discordgo.MessageEmbed {
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
