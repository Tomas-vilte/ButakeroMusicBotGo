package messaging

import (
	"github.com/Tomas-vilte/GoMusicBot/lambdas/message_processing/internal/logging"
	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

type DiscordMessenger interface {
	SendMessage(channelID string, embed *discordgo.MessageEmbed) error
	SendMessageToServers(embed *discordgo.MessageEmbed) error
}

type DiscordGoClient struct {
	Session DiscordSession
	Logger  logging.Logger
}

// NewDiscordGoClient crea una nueva instancia de DiscordGoClient.
func NewDiscordGoClient(session DiscordSession, logger logging.Logger) *DiscordGoClient {
	return &DiscordGoClient{
		Session: session,
		Logger:  logger,
	}
}

func (d *DiscordGoClient) SendMessage(channelID string, embed *discordgo.MessageEmbed) error {
	_, err := d.Session.ChannelMessageSendEmbed(channelID, embed)
	if err != nil {
		d.Logger.Error("Error al enviar mensaje al canal", zap.String("ID del canal:", channelID), zap.Error(err))
		return err
	}
	d.Logger.Info("Mensaje enviado al canal", zap.String("ID del canal:", channelID))
	return nil
}

// SendMessageToServers envía un mensaje a todos los servidores conectados.
func (d *DiscordGoClient) SendMessageToServers(embed *discordgo.MessageEmbed) error {
	// Obtener la lista de servidores
	guilds, err := d.Session.UserGuilds(0, "", "", true)
	if err != nil {
		d.Logger.Error("Error al obtener la lista de servidores", zap.Error(err))
		return err
	}

	// Iterar sobre cada servidor
	for _, guild := range guilds {
		// Verificar si el canal 'statusBot' existe
		channel, err := d.findStatusBotChannel(guild.ID)
		if err != nil {
			d.Logger.Error("Error al verificar el canal 'statusBot' en el servidor",
				zap.String("Servidor:", guild.Name),
				zap.Error(err),
			)
			continue
		}

		// Si el canal no existe, créalo
		if channel == nil {
			channel, err = d.createStatusBotChannel(guild.ID)
			if err != nil {
				d.Logger.Error("Error al crear el canal 'status-bot' en el servidor", zap.String("Servidor:", guild.Name), zap.Error(err))
				continue
			}
		}

		// Enviar el mensaje embed al canal 'statusBot'
		if err := d.SendMessage(channel.ID, embed); err != nil {
			d.Logger.Error("Error al enviar el mensaje a Discord en el servidor", zap.String("Servidor:", guild.Name), zap.Error(err))
		} else {
			d.Logger.Info("Mensaje enviado a Discord en el servidor",
				zap.String("Servidor:", guild.Name),
			)
		}
	}
	return nil
}

// findStatusBotChannel busca el canal 'statusBot' en el servidor especificado.
func (d *DiscordGoClient) findStatusBotChannel(guildID string) (*discordgo.Channel, error) {
	channels, err := d.Session.GuildChannels(guildID)
	if err != nil {
		d.Logger.Error("Error al obtener los channels", zap.Error(err))
		return nil, err
	}
	for _, channel := range channels {
		if channel.Name == "status-bot" && channel.Type == discordgo.ChannelTypeGuildText {
			return channel, nil
		}
	}
	return nil, nil // Canal no encontrado
}

// createStatusBotChannel crea el canal 'statusBot' dentro de una categoría en el servidor especificado.
func (d *DiscordGoClient) createStatusBotChannel(guildID string) (*discordgo.Channel, error) {
	// Crear la categoría si no existe
	category, err := d.createStatusBotCategory(guildID)
	if err != nil {
		d.Logger.Error("Error al crear la categoría 'Status Bot' en el servidor", zap.Error(err))
		return nil, err
	}

	// Crear el canal 'statusBot' dentro de la categoría
	newChannel, err := d.Session.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name:     "status-bot",
		Type:     discordgo.ChannelTypeGuildText,
		ParentID: category.ID,
	})
	if err != nil {
		d.Logger.Error("Error al crear el canal 'status-bot' en la categoría 'Status Bot' en el servidor", zap.Error(err))
		return nil, err
	}

	d.Logger.Info("Canal 'statusBot' creado en la categoría 'Status Bot' en el servidor",
		zap.String("ID del servidor:", guildID),
	)
	return newChannel, nil
}

// createStatusBotCategory crea la categoría 'Status Bot' en el servidor especificado si no existe.
func (d *DiscordGoClient) createStatusBotCategory(guildID string) (*discordgo.Channel, error) {
	// Buscar la categoría 'Status Bot'
	guildChannels, err := d.Session.GuildChannels(guildID)
	if err != nil {
		d.Logger.Error("Error al obtener los channels", zap.Error(err))
		return nil, err
	}
	for _, channel := range guildChannels {
		if channel.Name == "Status Bot" && channel.Type == discordgo.ChannelTypeGuildCategory {
			d.Logger.Error("La categoria ya existe en el servidor", zap.String("Nombre de la categoria:", channel.Name))
			return channel, nil // La categoría ya existe
		}
	}

	// Crear la categoría 'Status Bot' si no existe
	category, err := d.Session.GuildChannelCreate(guildID, "Status Bot", discordgo.ChannelTypeGuildCategory)
	if err != nil {
		d.Logger.Error("Error al crear la categoría 'Status Bot' en el servidor", zap.Error(err))
		return nil, err
	}

	d.Logger.Info("Categoría 'Status Bot' creada en el servidor", zap.String("ID del servidor:", guildID))
	return category, nil
}
