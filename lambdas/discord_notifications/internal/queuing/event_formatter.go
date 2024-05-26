package queuing

import (
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type EventFormatter interface {
	FormatEvent(event map[string]interface{}) (*discordgo.MessageEmbed, error)
}

type ReleaseEventFormatter struct{}

func (r *ReleaseEventFormatter) FormatEvent(event map[string]interface{}) (*discordgo.MessageEmbed, error) {
	release, ok := event["release"].(map[string]interface{})
	if !ok {
		return nil, errors.New("campo 'release' no encontrado o no es un mapa")
	}
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🎉 ¡Salio una nueva version kkkk: %s!", release["tag_name"]),
		Description: release["body"].(string),
		URL:         release["html_url"].(string),
		Color:       0x5865F2,
		Author: &discordgo.MessageEmbedAuthor{
			Name: "ButakeroMusicBotGo",
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://cdn.discordapp.com/attachments/1231503103279366207/1243293979471122453/github.png?ex=6650f33f&is=664fa1bf&hm=3ececa29784b9549657bd52bc18e375ffd2f840a97a95bcee0bca97d1445a01b&",
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Detalles de la Release",
				Value:  fmt.Sprintf("[**Aca esta los detalles por si les pinta**](%s)", release["html_url"]),
				Inline: false,
			},
		},
	}
	return embed, nil
}

type WorkflowActionEventFormatter struct{}

// FormatEvent formatea el evento de acción de workflow en un mensaje de Discord.
func (f *WorkflowActionEventFormatter) FormatEvent(event map[string]interface{}) (*discordgo.MessageEmbed, error) {
	workflowJob := event["workflow_job"].(map[string]interface{})

	embed := &discordgo.MessageEmbed{
		Title:       "🚀 Acción completada en el flujo de trabajo",
		Description: fmt.Sprintf("El flujo de trabajo **%s** ha completado una acción:", workflowJob["workflow_name"]),
		Color:       0x34a853, // Color verde para indicar éxito
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Estado", Value: fmt.Sprintf("**%s**", event["action"]), Inline: true},
			{Name: "Conclusión", Value: fmt.Sprintf("**%s**", workflowJob["conclusion"]), Inline: true},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "https://cdn.discordapp.com/attachments/1231503103279366207/1243293979471122453/github.png?ex=6650f33f&is=664fa1bf&hm=3ececa29784b9549657bd52bc18e375ffd2f840a97a95bcee0bca97d1445a01b&",
		},
	}

	// Agregar detalles adicionales si la acción fue exitosa
	if workflowJob["conclusion"] == "success" {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Detalles de la acción",
			Value:  fmt.Sprintf("[Ver detalles](%s)", workflowJob["html_url"]),
			Inline: false,
		})
	}

	// Agregar detalles del tiempo
	embed.Timestamp = workflowJob["completed_at"].(string)

	return embed, nil
}
