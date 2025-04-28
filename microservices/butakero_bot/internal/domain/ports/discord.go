package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
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
