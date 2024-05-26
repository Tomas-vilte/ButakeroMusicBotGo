package messaging

import (
	"github.com/bwmarrin/discordgo"
)

// DiscordSession define la interfaz para la sesión de Discord.
type DiscordSession interface {
	ChannelMessageSendEmbed(channelID string, embed *discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error)                // ChannelMessageSendEmbed envía un mensaje con un embed al canal especificado.
	UserGuilds(limit int, beforeID, afterID string, withCounts bool, options ...discordgo.RequestOption) (st []*discordgo.UserGuild, err error)             // UserGuilds obtiene los gremios a los que pertenece el usuario.
	GuildChannels(guildID string, options ...discordgo.RequestOption) (st []*discordgo.Channel, err error)                                                  // GuildChannels obtiene los canales de un gremio especificado.
	GuildChannelCreate(guildID, name string, ctype discordgo.ChannelType, options ...discordgo.RequestOption) (st *discordgo.Channel, err error)            // GuildChannelCreate crea un nuevo canal en un gremio.
	GuildChannelCreateComplex(guildID string, data discordgo.GuildChannelCreateData, options ...discordgo.RequestOption) (st *discordgo.Channel, err error) // GuildChannelCreateComplex crea un nuevo canal en un gremio con opciones adicionales.
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

// ChannelMessageSendEmbed envía un mensaje con un embed al canal especificado.
func (d *DiscordSessionImpl) ChannelMessageSendEmbed(channelID string, embed *discordgo.MessageEmbed, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	return d.session.ChannelMessageSendEmbed(channelID, embed, options...)
}

// UserGuilds obtiene los gremios a los que pertenece el usuario.
func (d *DiscordSessionImpl) UserGuilds(limit int, beforeID, afterID string, withCounts bool, options ...discordgo.RequestOption) (st []*discordgo.UserGuild, err error) {
	return d.session.UserGuilds(limit, beforeID, afterID, withCounts, options...)
}

// GuildChannels obtiene los canales de un gremio especificado.
func (d *DiscordSessionImpl) GuildChannels(guildID string, options ...discordgo.RequestOption) (st []*discordgo.Channel, err error) {
	return d.session.GuildChannels(guildID, options...)
}

// GuildChannelCreate crea un nuevo canal en un gremio.
func (d *DiscordSessionImpl) GuildChannelCreate(guildID, name string, ctype discordgo.ChannelType, options ...discordgo.RequestOption) (st *discordgo.Channel, err error) {
	return d.session.GuildChannelCreate(guildID, name, ctype, options...)
}

// GuildChannelCreateComplex crea un nuevo canal en un gremio con opciones adicionales.
func (d *DiscordSessionImpl) GuildChannelCreateComplex(guildID string, data discordgo.GuildChannelCreateData, options ...discordgo.RequestOption) (st *discordgo.Channel, err error) {
	return d.session.GuildChannelCreateComplex(guildID, data, options...)
}
