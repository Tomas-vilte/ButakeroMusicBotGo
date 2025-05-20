package command

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
)

type ResumeCommand struct {
	BaseCommand
	handler *CommandHandler
}

func NewResumeCommand(handler *CommandHandler, logger logging.Logger) Command {
	return &ResumeCommand{
		BaseCommand: BaseCommand{
			name:        "resume",
			description: "Reanudar la cancion actual",
			logger:      logger,
		},
		handler: handler,
	}
}

func (r *ResumeCommand) Handler() func(session *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(session *discordgo.Session, i *discordgo.InteractionCreate) {
		r.handler.ResumeSong(i)
	}
}
