package command

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
)

type PlayingCommand struct {
	BaseCommand
	handler *CommandHandler
}

func NewPlayingCommand(handler *CommandHandler, logger logging.Logger) Command {
	return &PlayingCommand{
		BaseCommand: BaseCommand{
			name:        "playing",
			description: "Mostrar la canción que se está reproduciendo actualmente",
			logger:      logger,
		},
		handler: handler,
	}
}

func (c *PlayingCommand) Handler() func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, ic *discordgo.InteractionCreate) {
		c.handler.GetPlayingSong(s, ic, nil)
	}
}
