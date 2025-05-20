package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

type (
	// GuildPlayer define las operaciones disponibles para manejar la reproducción de música en un servidor
	GuildPlayer interface {
		// Stop detiene la reproducción actual
		Stop(ctx context.Context) error

		// Pause pausa la reproducción actual
		Pause(ctx context.Context) error

		// Resume reanuda la reproducción pausada
		Resume(ctx context.Context) error

		// SkipSong salta a la siguiente canción en la lista
		SkipSong(ctx context.Context) error

		// AddSong agrega una nueva canción a la lista de reproducción
		AddSong(ctx context.Context, textChannelID, voiceChannelID *string, playedSong *entity.PlayedSong) error

		// RemoveSong elimina una canción de la lista por su posición
		RemoveSong(ctx context.Context, position int) (*entity.PlayedSong, error)

		// GetPlaylist obtiene la lista actual de canciones
		GetPlaylist(ctx context.Context) ([]*entity.PlayedSong, error)

		// GetPlayedSong obtiene la información de la canción que se está reproduciendo
		GetPlayedSong(ctx context.Context) (*entity.PlayedSong, error)

		// Close libera los recursos asociados al reproductor
		Close() error

		// MoveToVoiceChannel mueve el bot a un nuevo canal de voz
		MoveToVoiceChannel(ctx context.Context, newChannelID string) error
	}

	// GuildManager maneja los reproductores de música para diferentes servidores
	GuildManager interface {
		// CreateGuildPlayer crea un nuevo reproductor para un servidor
		CreateGuildPlayer(guildID string) (GuildPlayer, error)

		// RemoveGuildPlayer elimina el reproductor de un servidor
		RemoveGuildPlayer(guildID string) error

		// GetGuildPlayer obtiene el reproductor asociado a un servidor
		GetGuildPlayer(guildID string) (GuildPlayer, error)
	}
)
