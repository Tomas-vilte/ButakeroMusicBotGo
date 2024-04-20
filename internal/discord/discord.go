package discord

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

type DiscordClient struct {
	Session *discordgo.Session
}

func (d *DiscordClient) SearchGuildByChannelID(channelID string) (guildID string, err error) {
	channel, err := d.Session.Channel(channelID)
	if err != nil {
		return "", err
	}
	return channel.GuildID, nil
}

func (d *DiscordClient) SearchVoiceChannelByUserID(userID string) (voiceChannelID string) {
	for _, g := range d.Session.State.Guilds {
		for _, vs := range g.VoiceStates {
			if vs.UserID == userID {
				return vs.ChannelID
			}
		}
	}
	return ""
}

func (d *DiscordClient) SendChannelMessage(channelID, message string) {
	_, err := d.Session.ChannelMessageSend(channelID, message)
	if err != nil {
		log.Fatalln("Error en mandar el mensaje al canal"+"channelID"+channelID, "mensaje"+"error"+err.Error())
		return
	}
}

func (d *DiscordClient) JoinVoiceChannel(guildID, voiceChannelID string, mute, deafen bool) (*discordgo.VoiceConnection, error) {
	voiceConection, err := d.Session.ChannelVoiceJoin(guildID, voiceChannelID, mute, deafen)
	if err != nil {
		log.Fatalf("Error al unirse al canal de voz: %v", err)
		return nil, err
	}
	return voiceConection, nil
}
