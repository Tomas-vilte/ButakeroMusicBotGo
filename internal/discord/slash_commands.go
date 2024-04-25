package discord

import "github.com/bwmarrin/discordgo"

type SlashCommandRouter struct {
	commandPrefix string

	playHandler       func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)
	stopHandler       func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)
	listHandler       func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)
	skipHandler       func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)
	removeHandler     func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)
	playingNowHandler func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)
	djHandler         func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)

	addSongOrPlaylistHandler func(*discordgo.Session, *discordgo.InteractionCreate)
}

func NewSlashCommandRouter(commandPrefix string) *SlashCommandRouter {
	return &SlashCommandRouter{
		commandPrefix: commandPrefix,
	}
}

func (ch *SlashCommandRouter) PlayHandler(h func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)) *SlashCommandRouter {
	ch.playHandler = h
	return ch
}

func (ch *SlashCommandRouter) StopHandler(h func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)) *SlashCommandRouter {
	ch.stopHandler = h
	return ch
}

func (ch *SlashCommandRouter) SkipHandler(h func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)) *SlashCommandRouter {
	ch.skipHandler = h
	return ch
}

func (ch *SlashCommandRouter) ListHandler(h func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)) *SlashCommandRouter {
	ch.listHandler = h
	return ch
}

func (ch *SlashCommandRouter) RemoveHandler(h func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)) *SlashCommandRouter {
	ch.removeHandler = h
	return ch
}

func (ch *SlashCommandRouter) PlayingNowHandler(h func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)) *SlashCommandRouter {
	ch.playingNowHandler = h
	return ch
}

func (ch *SlashCommandRouter) DJHandler(h func(*discordgo.Session, *discordgo.InteractionCreate, *discordgo.ApplicationCommandInteractionDataOption)) *SlashCommandRouter {
	ch.djHandler = h
	return ch
}

func (ch *SlashCommandRouter) AddSongOrPlaylistHandler(h func(*discordgo.Session, *discordgo.InteractionCreate)) *SlashCommandRouter {
	ch.addSongOrPlaylistHandler = h
	return ch
}

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
			case "dj":
				ch.djHandler(s, ic, option)
			}
		},
	}
}

func (ch *SlashCommandRouter) GetComponentHandlers() map[string]func(*discordgo.Session, *discordgo.InteractionCreate) {
	return map[string]func(*discordgo.Session, *discordgo.InteractionCreate){
		"add_song_playlist": ch.addSongOrPlaylistHandler,
	}
}

func (ch *SlashCommandRouter) GetSlashCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        ch.commandPrefix,
			Description: "Air command",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "play",
					Description: "Add a song to the playlist",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "input",
							Description: "URL or name of the track",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "remove",
					Description: "Remove song from playlist",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "position",
							Description: "Position of the song in the playlist",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "skip",
					Description: "Skip the current song",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "stop",
					Description: "Stop playing and clear playlist",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list",
					Description: "List the playlist",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "playing",
					Description: "Get currently playing song",
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "dj",
					Description: "Create a playlist of songs",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "description",
							Description: "Description of the playlist",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionInteger,
							Name:        "length",
							Description: "Number of songs",
							Required:    false,
						},
					},
				},
			},
		},
	}
}
