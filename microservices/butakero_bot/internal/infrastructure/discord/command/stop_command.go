package command

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
)

type StopCommand struct {
	BaseCommand
	handler *CommandHandler
}

func NewStopCommand(handler *CommandHandler, logger logging.Logger) Command {
	return &StopCommand{
		BaseCommand: BaseCommand{
			name:        "stop",
			description: "Detener la reproducción y limpiar la lista de reproducción",
			logger:      logger,
		},
		handler: handler,
	}
}

func (c *StopCommand) Handler() func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, ic *discordgo.InteractionCreate) {
		c.handler.StopPlaying(s, ic, nil)
	}
}
