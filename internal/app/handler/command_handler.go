package handler

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/app/service"
	"github.com/bwmarrin/discordgo"
)

// CommandHandler representa un comando genérico.
type CommandHandler interface {
	RegisterCommands(s *discordgo.Session, guildID string) ([]*discordgo.ApplicationCommand, error)
	HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate)
}

// PingCommandHandler es una implementación concreta de CommandHandler para el comando de ping.
type PingCommandHandler struct {
	PingService service.PingService
}

func (h *PingCommandHandler) RegisterCommands(s *discordgo.Session) ([]*discordgo.ApplicationCommand, error) {
	command := &discordgo.ApplicationCommand{
		Name:        "ping",
		Description: "Activates the Ping Service.",
	}
	return []*discordgo.ApplicationCommand{command}, nil
}

func (h *PingCommandHandler) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.ApplicationCommandData().Name == "ping" {
		response := h.PingService.Ping()
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: response,
			},
		})
	}
}

//// NewCommandHandler crea una nueva instancia de CommandHandler.
//func NewCommandHandler() *CommandHandler {
//	return &CommandHandler{
//		commands: make(map[string]Command),
//	}
//}
//
//func (h *CommandHandler) RegisterCommand(name string, command Command) {
//	h.commands[name] = command
//}
//
//// Handle maneja los comandos del discord.
//func (h *CommandHandler) Handle(session *discordgo.Session, message *discordgo.MessageCreate) {
//	log.Println("Recibiendo un nuevo mensaje:", message.Content)
//	content := message.Content
//	if !strings.HasPrefix(content, "/") {
//		log.Println("El mensaje no es un comando.")
//		return
//	}
//
//	// Dividir el mensaje en comando y argumentos
//	parts := strings.Fields(content)
//	if len(parts) == 0 {
//		log.Println("Comando vacío.")
//		return
//	}
//	commandName := parts[0][1:] // Eliminar el prefijo "/"
//
//	command, ok := h.commands[commandName]
//	if !ok {
//		// Comando desconocido
//		log.Println("Comando desconocido:", commandName)
//		h.sendErrorMessage(session, message.ChannelID, "Comando desconocido: "+commandName)
//		return
//	}
//
//	if commandName == "play" {
//		musicChannelID := h.findMusicChannel(session, message.GuildID)
//		if musicChannelID == "" || message.ChannelID != musicChannelID {
//			log.Println("El comando de música solo puede ejecutarse en el canal de música o music.")
//			h.sendErrorMessage(session, message.ChannelID, "Los comandos de musica ponelo en un canal llamado musica forro, no uses el general")
//			return
//		}
//	}
//
//	log.Println("Manejando el comando:", commandName)
//	// Manejar el comando
//	command.Handle(session, message)
//}
//
//// findMusicChannel busca el canal de música en los canales del servidor.
//func (h *CommandHandler) findMusicChannel(session *discordgo.Session, guildID string) string {
//	channels, err := session.GuildChannels(guildID)
//	if err != nil {
//		log.Println("Error al obtener la lista de canales:", err)
//		return ""
//	}
//
//	for _, channel := range channels {
//		if channel.Name == "music" || channel.Name == "musica" {
//			return channel.ID
//		}
//	}
//	return ""
//}
//
//// sendErrorMessage envía un mensaje de error al canal especificado.
//func (h *CommandHandler) sendErrorMessage(session *discordgo.Session, channelID string, errorMessage string) {
//	_, err := session.ChannelMessageSend(channelID, errorMessage)
//	if err != nil {
//		log.Println("Error al enviar mensaje de error:", err)
//	}
//}
