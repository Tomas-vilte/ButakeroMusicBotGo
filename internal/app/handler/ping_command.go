package handler

import "github.com/bwmarrin/discordgo"

// PingCommand representa el comando de ping.
type PingCommand struct{}

func (p *PingCommand) Handle(session *discordgo.Session, message *discordgo.MessageCreate) {
	session.ChannelMessageSend(message.ChannelID, "Pong!")
}
