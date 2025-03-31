package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

type (
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

	DiscordMessenger interface {
		// RespondWithMessage responde a una interacción con un mensaje
		RespondWithMessage(interaction *entity.Interaction, message string) error
		// SendPlayStatus envía un mensaje embed de estado de reproducción
		SendPlayStatus(channelID string, playMsg *entity.PlayedSong) (messageID string, err error)
		// UpdatePlayStatus actualiza un mensaje de estado existente
		UpdatePlayStatus(channelID, messageID string, playMsg *entity.PlayedSong) error
		// SendText envía un mensaje de texto simple
		SendText(channelID, text string) error
		// Respond responde a una interacción con una respuesta estructurada
		Respond(interaction *entity.Interaction, response entity.InteractionResponse) error
		// CreateFollowupMessage crea un mensaje de seguimiento
		CreateFollowupMessage(interaction *entity.Interaction, params entity.WebhookParams) error
		// EditOriginalResponse edita la respuesta original
		EditOriginalResponse(interaction *entity.Interaction, params *entity.WebhookEdit) error
	}

	GuildManager interface {
		CreateGuildPlayer(guildID string) (GuildPlayer, error)
		RemoveGuildPlayer(guildID string) error
		GetGuildPlayer(guildID string) (GuildPlayer, error)
	}
)
