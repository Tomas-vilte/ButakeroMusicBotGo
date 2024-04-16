package handler

import (
	"github.com/bwmarrin/discordgo"
	"strings"
)

// Command representa un comando genérico.
type Command interface {
	Handle(s *discordgo.Session, m *discordgo.MessageCreate)
}

// CommandHandler maneja los comandos del bot.
type CommandHandler struct {
	commands map[string]Command
}

// NewCommandHandler crea una nueva instancia de CommandHandler.
func NewCommandHandler() *CommandHandler {
	return &CommandHandler{
		commands: make(map[string]Command),
	}
}

func (h *CommandHandler) RegisterCommand(name string, command Command) {
	h.commands[name] = command
}

// Handle maneja los comandos del bot.
func (h *CommandHandler) Handle(session *discordgo.Session, message *discordgo.MessageCreate) {
	content := message.Content
	if !strings.HasPrefix(content, "!") {
		return
	}

	// Dividir el mensaje en comando y argumentos
	partes := strings.Fields(content)
	commandName := partes[0][1:] // eliminar el prefijo !

	command, ok := h.commands[commandName]
	if !ok {
		// comando no registrado
		return
	}

	// Manejar el comando
	command.Handle(session, message)
}
