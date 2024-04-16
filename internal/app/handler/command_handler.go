package handler

import (
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/app/service"
	"github.com/bwmarrin/discordgo"
)

// CommandHandler es responsable de manejar los comandos del bot.
type CommandHandler struct {
	audioService service.AudioService
}

// NewCommandHandler crea una nueva instancia de CommandHandler.
func NewCommandHandler(audioService service.AudioService) *CommandHandler {
	return &CommandHandler{audioService: audioService}
}

// Handle maneja los comandos del bot.
func (h *CommandHandler) Handle(session *discordgo.Session, message *discordgo.MessageCreate) {
	content := message.Content
	if content[0] != '!' {
		return
	}

	command, args := parseCommand(content)

	switch command {
	case "!play":
		h.handlePlayCommand(session, message, args)
	default:
		fmt.Println("Comando desconocido")
	}
}

// handlePlayCommand maneja el comando de reproducción de música.
func (h *CommandHandler) handlePlayCommand(session *discordgo.Session, message *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		session.ChannelMessageSend(message.ChannelID, "Por favor, proporciona una URL para reproducir.")
		return
	}
	err := h.audioService.PlayAudio(session, message, args[0])
	if err != nil {
		session.ChannelMessageSend(message.ChannelID, "Error al reproducir audio: "+err.Error())
	}
}

// Función para dividir el comando y los argumentos
func parseCommand(content string) (string, []string) {
	command := ""
	var args []string

	for i, char := range content {
		if i == 0 && char == '!' {
			continue
		}
		if char == ' ' {
			command = content[1:i]
			args = append(args, content[i+1:])
			break
		}
	}
	return command, args
}
