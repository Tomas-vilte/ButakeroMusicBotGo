package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
)

type (
	// ExternalSongService define una interfaz para servicios externos de canciones.
	ExternalSongService interface {
		// RequestDownload solicita la descarga de una canción por su nombre o url.
		// Devuelve una respuesta de descarga o un error si ocurre algún problema.
		RequestDownload(ctx context.Context, songName, providerType string) (*entity.DownloadResponse, error)
	}

	SongService interface {
		GetOrDownloadSong(ctx context.Context, songInput, providerType string) (*entity.DiscordEntity, error)
	}
)
