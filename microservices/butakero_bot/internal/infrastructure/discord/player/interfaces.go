package player

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

// PlaybackHandler maneja la reproducción de audio
type PlaybackHandler interface {
	Play(ctx context.Context, song *entity.PlayedSong, textChannel string) error
	Pause(ctx context.Context) error
	Resume(ctx context.Context) error
	Stop(ctx context.Context)
	CurrentState() PlayerState
}

// PlaylistHandler maneja operaciones con la lista de reproducción
type PlaylistHandler interface {
	AddSong(ctx context.Context, song *entity.PlayedSong) error
	RemoveSong(ctx context.Context, position int) (*entity.DiscordEntity, error)
	GetPlaylist(ctx context.Context) ([]string, error)
	ClearPlaylist(ctx context.Context) error
	GetNextSong(ctx context.Context) (*entity.PlayedSong, error)
}
