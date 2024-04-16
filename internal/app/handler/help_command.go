package handler

import "github.com/bwmarrin/discordgo"

type HelpCommand struct{}

func (h *HelpCommand) Handle(session *discordgo.Session, message *discordgo.MessageCreate) {
	session.ChannelMessageSend(message.ChannelID, "¡Bienvenido! Aquí tienes una lista de comandos disponibles:\n!ping - Responde con 'Pong!'\n!help - Muestra este mensaje de ayuda")
}
