package discord

import (
	"github.com/bwmarrin/discordgo"
)

func getMemberName(member *discordgo.Member) string {
	if member.Nick != "" {
		return member.Nick
	}
	return member.User.Username
}
