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
)
