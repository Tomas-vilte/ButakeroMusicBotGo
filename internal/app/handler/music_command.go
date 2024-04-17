package handler

import (
	"github.com/bwmarrin/discordgo"
)

type MusicCommand struct{}

func (m *MusicCommand) Handle(session *discordgo.Session, message *discordgo.MessageCreate) {
	// TODO: implementar logica para reproducir musica
	session.ChannelMessageSend(message.ChannelID, "Reproduciendo música...")
}
