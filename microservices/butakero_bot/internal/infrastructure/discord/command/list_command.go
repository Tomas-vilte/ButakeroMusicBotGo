package command

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
)

type ListCommand struct {
	BaseCommand
	handler *CommandHandler
}

func NewListCommand(handler *CommandHandler, logger logging.Logger) Command {
	return &ListCommand{
		BaseCommand: BaseCommand{
			name:        "list",
			description: "Mostrar la lista de reproducción actual",
			logger:      logger,
		},
		handler: handler,
	}
}

func (c *ListCommand) Handler() func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, ic *discordgo.InteractionCreate) {
		c.handler.ListPlaylist(s, ic, nil)
	}
}
