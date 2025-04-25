package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model/discord"
	"io"
)

type (
	GuildPlayer interface {
		Stop(ctx context.Context) error
		Pause(ctx context.Context) error
		Resume(ctx context.Context) error
		SkipSong(ctx context.Context)
		AddSong(ctx context.Context, textChannelID, voiceChannelID *string, playedSong *entity.PlayedSong) error
		RemoveSong(ctx context.Context, position int) (*entity.DiscordEntity, error)
		GetPlaylist(ctx context.Context) ([]string, error)
		GetPlayedSong(ctx context.Context) (*entity.PlayedSong, error)
		StateStorage() PlayerStateStorage
		JoinVoiceChannel(ctx context.Context, channelID string) error
	}

	DiscordMessenger interface {
		// RespondWithMessage responde a una interacción con un mensaje
		RespondWithMessage(interaction *discord.Interaction, message string) error
		// SendPlayStatus envía un mensaje embed de estado de reproducción
		SendPlayStatus(channelID string, playMsg *entity.PlayedSong) (messageID string, err error)
		// UpdatePlayStatus actualiza un mensaje de estado existente
		UpdatePlayStatus(channelID, messageID string, playMsg *entity.PlayedSong) error
		// SendText envía un mensaje de texto simple
		SendText(channelID, text string) error
		// Respond responde a una interacción con una respuesta estructurada
		Respond(interaction *discord.Interaction, response discord.InteractionResponse) error
		// CreateFollowupMessage crea un mensaje de seguimiento
		CreateFollowupMessage(interaction *discord.Interaction, params discord.WebhookParams) error
		// EditOriginalResponse edita la respuesta original
		EditOriginalResponse(interaction *discord.Interaction, params *discord.WebhookEdit) error
	}

	GuildManager interface {
		CreateGuildPlayer(guildID string) (GuildPlayer, error)
		RemoveGuildPlayer(guildID string) error
		GetGuildPlayer(guildID string) (GuildPlayer, error)
	}

	// VoiceSession define una interfaz para manejar sesiones de voz.
	VoiceSession interface {
		// JoinVoiceChannel une a un canal de voz especificado por channelID.
		JoinVoiceChannel(ctx context.Context, channelID string) error
		// LeaveVoiceChannel deja el canal de voz actual.
		LeaveVoiceChannel(ctx context.Context) error
		// SendAudio envía audio a través de la sesión de voz.
		SendAudio(ctx context.Context, reader io.ReadCloser) error
		// Pause pausa la sesión de voz.
		Pause()
		// Resume reanuda la sesión de voz.
		Resume()
	}
)
