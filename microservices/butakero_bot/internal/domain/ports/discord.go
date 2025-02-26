package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/bwmarrin/discordgo"
)

// DiscordMessenger define todas las operaciones de mensajería relacionadas con Discord.
type (
	DiscordMessenger interface {
		// RespondToInteraction Responde a una interacción (comando slash, botón, etc.) con un embed.
		RespondToInteraction(interaction *discordgo.Interaction, embed *discordgo.MessageEmbed) error
		// SendPlayStatus Envía un mensaje embed de estado de reproducción (ej: "Ahora sonando").
		SendPlayStatus(channelID string, playMsg *entity.PlayedSong) (messageID string, err error)
		// UpdatePlayStatus Actualiza un mensaje de estado de reproducción existente.
		UpdatePlayStatus(channelID, messageID string, playMsg *entity.PlayedSong) error
		// SendText Envía un mensaje de texto simple a un canal.
		SendText(channelID, text string) error
	}

	InteractionHandler interface {
		PlaySong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption)
		StopPlaying(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption)
		SkipSong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption)
		ListPlaylist(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption)
		RemoveSong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption)
		GetPlayingSong(s *discordgo.Session, ic *discordgo.InteractionCreate, opt *discordgo.ApplicationCommandInteractionDataOption)
		Ready(s *discordgo.Session, event *discordgo.Ready)
		GuildCreate(ctx context.Context, s *discordgo.Session, event *discordgo.GuildCreate)
		GuildDelete(s *discordgo.Session, event *discordgo.GuildDelete)
		RegisterEventHandlers(s *discordgo.Session, ctx context.Context)
	}

	GuildPlayer interface {
		Run(ctx context.Context) error
		Stop() error
		SkipSong()
		Close() error
		AddSong(textChannelID, voiceChannelID *string, songs ...*entity.Song) error
		RemoveSong(position int) (*entity.Song, error)
		GetPlaylist() ([]string, error)
		GetPlayedSong() (*entity.PlayedSong, error)
	}
)
