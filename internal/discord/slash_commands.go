package discord

import "github.com/bwmarrin/discordgo"

// SlashCommandRouter enruta los comandos de barra oblicua en Discord.
type SlashCommandRouter struct {
	commandPrefix            string
	playHandler              func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)
	stopHandler              func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)
	listHandler              func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)
	skipHandler              func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)
	removeHandler            func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)
	playingNowHandler        func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)
	addSongOrPlaylistHandler func(*discordgo.Session, *discordgo.InteractionCreate)
}

// NewSlashCommandRouter crea una nueva instancia de SlashCommandRouter con el prefijo de comando especificado.
func NewSlashCommandRouter(commandPrefix string) *SlashCommandRouter {
	return &SlashCommandRouter{
		commandPrefix: commandPrefix,
	}
}

// PlayHandler establece el manejador para el comando "play".
func (ch *SlashCommandRouter) PlayHandler(h func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)) *SlashCommandRouter {
	ch.playHandler = h
	return ch
}

// StopHandler establece el manejador para el comando "stop".
func (ch *SlashCommandRouter) StopHandler(h func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)) *SlashCommandRouter {
	ch.stopHandler = h
	return ch
}

// SkipHandler establece el manejador para el comando "skip".
func (ch *SlashCommandRouter) SkipHandler(h func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)) *SlashCommandRouter {
	ch.skipHandler = h
	return ch
}

// ListHandler establece el manejador para el comando "list".
func (ch *SlashCommandRouter) ListHandler(h func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)) *SlashCommandRouter {
	ch.listHandler = h
	return ch
}

// RemoveHandler establece el manejador para el comando "remove".
func (ch *SlashCommandRouter) RemoveHandler(h func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)) *SlashCommandRouter {
	ch.removeHandler = h
	return ch
}

// PlayingNowHandler establece el manejador para el comando "playing".
func (ch *SlashCommandRouter) PlayingNowHandler(h func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)) *SlashCommandRouter {
	ch.playingNowHandler = h
	return ch
}

// AddSongOrPlaylistHandler establece el manejador para el comando "add_song_playlist".
func (ch *SlashCommandRouter) AddSongOrPlaylistHandler(h func(*discordgo.Session, *discordgo.InteractionCreate)) *SlashCommandRouter {
	ch.addSongOrPlaylistHandler = h
	return ch
}

// GetCommandHandlers devuelve los manejadores de los comandos de barra oblicua.
func (ch *SlashCommandRouter) GetCommandHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate) {
	return map[string]func(*discordgo.Session, *discordgo.InteractionCreate){
		ch.commandPrefix: func(s *discordgo.Session, ic *discordgo.InteractionCreate) {
			options := ic.ApplicationCommandData().Options
			option := options[0]

			switch option.Name {
			case "play":
				ch.playHandler(s, ic, option)
			case "stop":
				ch.stopHandler(s, ic, option)
			case "list":
				ch.listHandler(s, ic, option)
			case "skip":
				ch.skipHandler(s, ic, option)
			case "remove":
				ch.removeHandler(s, ic, option)
			case "playing":
				ch.playingNowHandler(s, ic, option)
			}
		},
	}
}

// GetComponentHandlers devuelve los manejadores de los componentes.
func (ch *SlashCommandRouter) GetComponentHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate) {
	return map[string]func(*discordgo.Session, *discordgo.InteractionCreate){
		"add_song_playlist": ch.addSongOrPlaylistHandler,
	}
}

// GetSlashCommands devuelve los comandos de barra oblicua.
func (ch *SlashCommandRouter) GetSlashCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        ch.commandPrefix,
			Description: "Comando de aire",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "play",
					Description: "Agregar una canción a la lista de reproducción",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "input",
							Description: "URL o nombre de la pista",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "Eliminar canción de la lista de reproducción",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "position",
							Description: "Posición de la canción en la lista de reproducción",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "skip",
					Description: "Saltar la canción actual",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "stop",
					Description: "Detener la reproducción y limpiar la lista de reproducción",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list",
					Description: "Listar la lista de reproducción",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "playing",
					Description: "Obtener la canción que se está reproduciendo actualmente",
				},
			},
		},
	}
}
