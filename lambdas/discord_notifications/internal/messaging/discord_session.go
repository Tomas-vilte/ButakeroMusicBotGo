package messaging

import (
	"github.com/bwmarrin/discordgo"
)

type DiscordSession interface {
	ChannelMessageSendEmbed(channelID string, embed *discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error)
	UserGuilds(limit int, beforeID, afterID string, withCounts bool, options ...discordgo.RequestOption) (st []*discordgo.UserGuild, err error)
	GuildChannels(guildID string, options ...discordgo.RequestOption) (st []*discordgo.Channel, err error)
	GuildChannelCreate(guildID, name string, ctype discordgo.ChannelType, options ...discordgo.RequestOption) (st *discordgo.Channel, err error)
	GuildChannelCreateComplex(guildID string, data discordgo.GuildChannelCreateData, options ...discordgo.RequestOption) (st *discordgo.Channel, err error)
}

// DiscordSessionImpl es una implementación de la interfaz DiscordSession que se conecta a Discord utilizando discordgo.Session.
type DiscordSessionImpl struct {
	session *discordgo.Session
}

// NewDiscordSessionImpl crea una nueva instancia de DiscordSessionImpl y la conecta a Discord.
func NewDiscordSessionImpl(token string) (*DiscordSessionImpl, error) {
	discordSession, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	err = discordSession.Open()
	if err != nil {
		return nil, err
	}

	return &DiscordSessionImpl{
		session: discordSession,
	}, nil
}

// ChannelMessageSendEmbed Implementación de los métodos de la interfaz DiscordSession para DiscordSessionImpl.
func (d *DiscordSessionImpl) ChannelMessageSendEmbed(channelID string, embed *discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	return d.session.ChannelMessageSendEmbed(channelID, embed, options...)
}

func (d *DiscordSessionImpl) UserGuilds(limit int, beforeID, afterID string, withCounts bool, options ...discordgo.RequestOption) (st []*discordgo.UserGuild, err error) {
	return d.session.UserGuilds(limit, beforeID, afterID, withCounts, options...)
}

func (d *DiscordSessionImpl) GuildChannels(guildID string, options ...discordgo.RequestOption) (st []*discordgo.Channel, err error) {
	return d.session.GuildChannels(guildID, options...)
}

func (d *DiscordSessionImpl) GuildChannelCreate(guildID, name string, ctype discordgo.ChannelType, options ...discordgo.RequestOption) (st *discordgo.Channel, err error) {
	return d.session.GuildChannelCreate(guildID, name, ctype, options...)
}

func (d *DiscordSessionImpl) GuildChannelCreateComplex(guildID string, data discordgo.GuildChannelCreateData, options ...discordgo.RequestOption) (st *discordgo.Channel, err error) {
	return d.session.GuildChannelCreateComplex(guildID, data, options...)
}
