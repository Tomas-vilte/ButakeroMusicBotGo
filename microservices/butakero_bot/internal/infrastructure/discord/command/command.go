package command

import (
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/shared/logging"
	"github.com/bwmarrin/discordgo"
)

type Command interface {
	Name() string
	Description() string
	Options() []*discordgo.ApplicationCommandOption
	Handler() func(*discordgo.Session, *discordgo.InteractionCreate)
}

type BaseCommand struct {
	name        string
	description string
	options     []*discordgo.ApplicationCommandOption
	handler     func(*discordgo.Session, *discordgo.InteractionCreate)
	logger      logging.Logger
}

func (c *BaseCommand) Name() string {
	return c.name
}

func (c *BaseCommand) Description() string {
	return c.description
}

func (c *BaseCommand) Options() []*discordgo.ApplicationCommandOption {
	return c.options
}

func (c *BaseCommand) Handler() func(*discordgo.Session, *discordgo.InteractionCreate) {
	return c.handler
}
