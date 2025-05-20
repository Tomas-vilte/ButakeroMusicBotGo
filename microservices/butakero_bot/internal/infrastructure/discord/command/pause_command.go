package command

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
)

type PauseCommand struct {
	BaseCommand
	handler *CommandHandler
}

func NewPauseCommand(handler *CommandHandler, logger logging.Logger) Command {
	return &PauseCommand{
		BaseCommand: BaseCommand{
			name:        "pause",
			description: "Pausa la reproduccion actual",
			logger:      logger,
		},
		handler: handler,
	}
}

func (p *PauseCommand) Handler() func(session *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(session *discordgo.Session, i *discordgo.InteractionCreate) {
		p.handler.PauseSong(i)
	}
}
