package command

import "github.com/bwmarrin/discordgo"

type CommandRegistry struct {
	commands map[string]Command
}

func NewCommandRegistry() *CommandRegistry {
	return &CommandRegistry{
		commands: make(map[string]Command),
	}
}

func (r *CommandRegistry) Register(command Command) {
	r.commands[command.Name()] = command
}

func (r *CommandRegistry) GetCommands() []*discordgo.ApplicationCommand {
	var appCommands []*discordgo.ApplicationCommand

	for _, cmd := range r.commands {
		appCommand := &discordgo.ApplicationCommand{
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Options:     cmd.Options(),
		}
		appCommands = append(appCommands, appCommand)
	}

	return appCommands
}

func (r *CommandRegistry) GetCommandHandlers() map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	handlers := make(map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate))

	for name, cmd := range r.commands {
		handlers[name] = cmd.Handler()
	}
	return handlers
}
