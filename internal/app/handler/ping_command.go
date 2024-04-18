package handler

import (
	"github.com/bwmarrin/discordgo"
)

// PingCommand representa el comando de ping.
type PingCommand struct{}

func (p *PingCommand) Handle(session *discordgo.Session, msg *discordgo.MessageCreate) {
	session.ChannelMessageSend(msg.ChannelID, "Pong!")
}
