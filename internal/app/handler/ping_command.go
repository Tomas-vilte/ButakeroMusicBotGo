package handler

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/message"
	"github.com/bwmarrin/discordgo"
	"log"
)

// PingCommand representa el comando de ping.
type PingCommand struct{}

func (p *PingCommand) Handle(session *discordgo.Session, msg *discordgo.MessageCreate) {
	session.ChannelMessageSend(msg.ChannelID, "Pong!")

	// crear mensaje de info del comando
	messageResponse := &message.CommandInfoMessage{
		CommandName: "ping",
		Description: "Responde con 'Pong!' para comprobar si el bot está en línea.",
		Usage:       "!ping",
		Args:        "Ninguno",
		Permissions: "Ninguno",
		Aliases:     []string{"p"},
		Examples:    []string{"!ping"},
	}
	err := messageResponse.SendMessage(session, msg.ChannelID)
	if err != nil {
		log.Printf("Hubo un error %v\n", err)
		session.ChannelMessageSend(msg.ChannelID, "Ocurrió un error al obtener la información del comando: "+err.Error())
	}
}
