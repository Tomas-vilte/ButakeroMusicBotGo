package command

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
)

type RemoveCommand struct {
	BaseCommand
	handler *CommandHandler
}

func NewRemoveCommand(handler *CommandHandler, logger logging.Logger) Command {
	return &RemoveCommand{
		BaseCommand: BaseCommand{
			name:        "remove",
			description: "Eliminar una canción de la lista de reproducción",
			options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "position",
					Description: "Posición de la canción a eliminar",
					Required:    true,
				},
			},
			logger: logger,
		},
		handler: handler,
	}
}

func (c *RemoveCommand) Handler() func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, ic *discordgo.InteractionCreate) {
		if len(ic.ApplicationCommandData().Options) == 0 {
			c.logger.Error("No se proporcionó posición para eliminar")
			return
		}
		opt := ic.ApplicationCommandData().Options[0]
		c.handler.RemoveSong(s, ic, opt)
	}
}
