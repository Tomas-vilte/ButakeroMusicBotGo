package handler

import (
	"github.com/bwmarrin/discordgo"
	"log"
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
	log.Println("Recibiendo un nuevo mensaje:", message.Content)
	content := message.Content
	if !strings.HasPrefix(content, "!") {
		log.Println("El mensaje no es un comando.")
		return
	}

	// Dividir el mensaje en comando y argumentos
	parts := strings.Fields(content)
	commandName := parts[0][1:] // Eliminar el prefijo "!"

	command, ok := h.commands[commandName]
	if !ok {
		// commando desconocido
		log.Println("Comando desconocido:", commandName)
		h.sendErrorMessage(session, message.ChannelID, "Comando desconocido: "+commandName)
		return
	}

	// Validar que el mensaje se envió desde el canal correcto para el comando de música
	if commandName == "play" {
		musicChannelID := h.findMusicChannel(session, message.GuildID)
		if musicChannelID == "" || message.ChannelID != musicChannelID {
			log.Println("El comando de música solo puede ejecutarse en el canal de música o music.")
			h.sendErrorMessage(session, message.ChannelID, "Los comandos de musica ponelo en un canal llamado musica forro, no uses el general")
			return
		}
	}

	log.Println("Manejando el comando:", commandName)
	// Manejar el comando
	command.Handle(session, message)
}

// findMusicChannel busca el canal de música en los canales del servidor.
func (h *CommandHandler) findMusicChannel(session *discordgo.Session, guildID string) string {
	channels, err := session.GuildChannels(guildID)
	if err != nil {
		log.Println("Error al obtener la lista de canales:", err)
		return ""
	}

	for _, channel := range channels {
		if channel.Name == "music" || channel.Name == "musica" {
			return channel.ID
		}
	}
	return ""
}

// sendErrorMessage envía un mensaje de error al canal especificado.
func (h *CommandHandler) sendErrorMessage(session *discordgo.Session, channelID string, errorMessage string) {
	_, err := session.ChannelMessageSend(channelID, errorMessage)
	if err != nil {
		log.Println("Error al enviar mensaje de error:", err)
	}
}
