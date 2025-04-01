package player

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/ports"
)

// PlaybackHandler maneja la reproducción de audio
type PlaybackHandler interface {
	Play(ctx context.Context, song *entity.PlayedSong, textChannel string) error
	Pause() error
	Resume() error
	Stop()
	CurrentState() PlayerState
	GetVoiceSession() ports.VoiceSession
}

// PlaylistHandler maneja operaciones con la lista de reproducción
type PlaylistHandler interface {
	AddSong(song *entity.PlayedSong) error
	RemoveSong(position int) (*entity.DiscordEntity, error)
	GetPlaylist() ([]string, error)
	ClearPlaylist() error
	GetNextSong() (*entity.PlayedSong, error)
}
