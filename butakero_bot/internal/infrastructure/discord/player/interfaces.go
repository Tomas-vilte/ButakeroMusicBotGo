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
