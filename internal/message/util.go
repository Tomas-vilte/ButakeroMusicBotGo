package message

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"strings"
)

type Messenger interface {
	SendMessage(session *discordgo.Session, channelID string) error
}

type ErrorMessage struct {
	Message string
}

func (e *ErrorMessage) SendMessage(session *discordgo.Session, channelID string) error {
	embed := &discordgo.MessageEmbed{
		Title:       "Error",
		Description: e.Message,
		Color:       0xFF0000,
	}

	_, err := session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		log.Println("error al enviar mensaje de error: %w", err)
	}
	return nil
}

type CommandInfoMessage struct {
	CommandName string
	Description string
	Usage       string
	Args        string
	Permissions string
	Aliases     []string
	Examples    []string
}

func (c *CommandInfoMessage) SendMessage(session *discordgo.Session, channelID string) error {
	embed := &discordgo.MessageEmbed{
		Title:       "Información del comando: " + c.CommandName,
		Description: c.Description,
		Color:       0x00FF00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Uso", Value: "```" + c.Usage + "```"},
			{Name: "Argumentos", Value: c.Args},
			{Name: "Permisos requeridos", Value: c.Permissions},
		},
	}
	if len(c.Aliases) > 0 {
		aliasesStr := fmt.Sprintf("```%s```", strings.Join(c.Aliases, ", "))
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Alias", Value: aliasesStr})
	}

	if len(c.Examples) > 0 {
		examplesStr := fmt.Sprintf("```%s```", strings.Join(c.Examples, "\n"))
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{Name: "Ejemplos", Value: examplesStr})
	}

	_, err := session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		log.Println("error al enviar mensaje de error: %w", err)
		return err
	}
	return nil
}
