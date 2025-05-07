package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

// PlayerStateStorage define métodos para el almacenamiento y manipulación del estado del reproductor de música.
type (
	PlayerStateStorage interface {
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

	// PlaylistStorage define métodos para el almacenamiento y manipulación de la lista de reproducción de canciones.
	PlaylistStorage interface {
		// AppendTrack agrega una canción al final de la lista de reproducción.
		AppendTrack(ctx context.Context, track *entity.PlayedSong) error
		// RemoveTrack elimina una canción de la lista de reproducción por su posición.
		RemoveTrack(ctx context.Context, position int) (*entity.PlayedSong, error)
		// ClearPlaylist elimina todas las canciones de la lista de reproducción.
		ClearPlaylist(ctx context.Context) error
		// GetAllTracks devuelve todas las canciones en la lista de reproducción.
		GetAllTracks(ctx context.Context) ([]*entity.PlayedSong, error)
		// PopNextTrack elimina y devuelve la primera canción de la lista de reproducción.
		PopNextTrack(ctx context.Context) (*entity.PlayedSong, error)
	}

	// InteractionStorage define la interfaz para el almacenamiento de interacciones.
	InteractionStorage interface {
		// SaveSongList guarda una lista de canciones asociada a un canal.
		SaveSongList(channelID string, list []*entity.DiscordEntity)
		// GetSongList obtiene la lista de canciones asociada a un canal.
		GetSongList(channelID string) []*entity.DiscordEntity
		// DeleteSongList elimina la lista de canciones asociada a un canal.
		DeleteSongList(channelID string)
	}
)
