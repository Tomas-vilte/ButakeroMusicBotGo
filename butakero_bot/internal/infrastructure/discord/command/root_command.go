package command

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type RootCommand struct {
	BaseCommand
	subCommands []Command
	prefix      string
}

func NewRootCommand(prefix string, subCommands []Command, logger logging.Logger) Command {
	return &RootCommand{
		BaseCommand: BaseCommand{
			name:        prefix,
			description: "Comandos principales del bot de música",
			logger:      logger,
		},
		subCommands: subCommands,
		prefix:      prefix,
	}
}

func (c *RootCommand) Options() []*discordgo.ApplicationCommandOption {
	var options []*discordgo.ApplicationCommandOption
	for _, cmd := range c.subCommands {
		options = append(options, &discordgo.ApplicationCommandOption{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Options:     cmd.Options(),
		})
	}
	return options
}

func (c *RootCommand) Handler() func(*discordgo.Session, *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, ic *discordgo.InteractionCreate) {
		if len(ic.ApplicationCommandData().Options) == 0 {
			c.logger.Error("No se proporcionó subcomando")
			return
		}

		subCmdName := ic.ApplicationCommandData().Options[0].Name
		for _, cmd := range c.subCommands {
			if cmd.Name() == subCmdName {
				cmd.Handler()(s, ic)
				return
			}
		}
		c.logger.Error("Subcomando no encontrado", zap.String("subcomando", subCmdName))
	}
}
