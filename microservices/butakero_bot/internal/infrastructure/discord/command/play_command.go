package command

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
)

type PlayCommand struct {
	BaseCommand
	handler *CommandHandler
}

func NewPlayCommand(handler *CommandHandler, logger logging.Logger) *PlayCommand {
	return &PlayCommand{
		BaseCommand: BaseCommand{
			name:        "play",
			description: "Agregar una canción a la lista de reproducción",
			options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "input",
					Description: "URL o nombre de la pista",
					Required:    true,
				},
			},
			logger: logger,
		},
		handler: handler,
	}
}

func (c *PlayCommand) Handler() func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, ic *discordgo.InteractionCreate) {
		if len(ic.ApplicationCommandData().Options) == 0 {
			c.logger.Error("No se proporcionaron opciones para el comando play")
			return
		}
		opt := ic.ApplicationCommandData().Options[0]
		c.handler.PlaySong(s, ic, opt)
	}
}
