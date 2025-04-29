package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
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
		Close() error
	}

	GuildManager interface {
		CreateGuildPlayer(guildID string) (GuildPlayer, error)
		RemoveGuildPlayer(guildID string) error
		GetGuildPlayer(guildID string) (GuildPlayer, error)
	}
)
