package command

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
)

type SkipCommand struct {
	BaseCommand
	handler *CommandHandler
}

func NewSkipCommand(handler *CommandHandler, logger logging.Logger) Command {
	return &SkipCommand{
		BaseCommand: BaseCommand{
			name:        "skip",
			description: "Saltar la canción actual",
			logger:      logger,
		},
		handler: handler,
	}
}

func (c *SkipCommand) Handler() func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, ic *discordgo.InteractionCreate) {
		c.handler.SkipSong(ic)
	}
}
