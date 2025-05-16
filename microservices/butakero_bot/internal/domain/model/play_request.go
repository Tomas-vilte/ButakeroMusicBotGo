package model

import "context"

type PlayRequestData struct {
	Ctx             context.Context
	GuildID         string
	ChannelID       string
	VoiceChannelID  string
	UserID          string
	SongInput       string
	OriginalMsgID   string
	RequestedByName string
}
