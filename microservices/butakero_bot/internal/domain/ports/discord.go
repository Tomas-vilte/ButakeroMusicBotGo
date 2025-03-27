package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/bwmarrin/discordgo"
)

// DiscordMessenger define todas las operaciones de mensajería relacionadas con Discord.
type (
	DiscordMessenger interface {
		// RespondWithMessage Responde a una interacción (comando slash, botón, etc.) con un mensaje.
		RespondWithMessage(interaction *discordgo.Interaction, message string) error
		// SendPlayStatus Envía un mensaje embed de estado de reproducción (ej: "Ahora sonando").
		SendPlayStatus(channelID string, playMsg *entity.PlayedSong) (messageID string, err error)
		// UpdatePlayStatus Actualiza un mensaje de estado de reproducción existente.
		UpdatePlayStatus(channelID, messageID string, playMsg *entity.PlayedSong) error
		// SendText Envía un mensaje de texto simple a un canal.
		SendText(channelID, text string) error
		Respond(interaction *discordgo.Interaction, response discordgo.InteractionResponse) error
		CreateFollowupMessage(interaction *discordgo.Interaction, params discordgo.WebhookParams) error
		EditOriginalResponse(interaction *discordgo.Interaction, params *discordgo.WebhookEdit) error
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
		AddSong(textChannelID, voiceChannelID *string, playedSong *entity.PlayedSong) error
		RemoveSong(position int) (*entity.DiscordEntity, error)
		GetPlaylist() ([]string, error)
		GetPlayedSong() (*entity.PlayedSong, error)
		Session() VoiceSession
		StateStorage() StateStorage
	}
)
