package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

// PlayerStateStorage define métodos para el almacenamiento y manipulación del estado del reproductor de música.
type PlayerStateStorage interface {
	// GetCurrentTrack devuelve la canción que se está reproduciendo actualmente.
	GetCurrentTrack(ctx context.Context) (*entity.PlayedSong, error)
	// SetCurrentTrack establece la canción que se está reproduciendo actualmente.
	SetCurrentTrack(ctx context.Context, track *entity.PlayedSong) error
	// GetVoiceChannelID devuelve el ID del canal de voz actual.
	GetVoiceChannelID(ctx context.Context) (string, error)
	// SetVoiceChannelID establece el ID del canal de voz actual.
	SetVoiceChannelID(ctx context.Context, channelID string) error
	// GetTextChannelID devuelve el ID del canal de texto actual.
	GetTextChannelID(ctx context.Context) (string, error)
	// SetTextChannelID establece el ID del canal de texto actual.
	SetTextChannelID(ctx context.Context, channelID string) error
}
