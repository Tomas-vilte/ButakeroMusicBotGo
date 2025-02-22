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
		RequestDownload(ctx context.Context, songName string) (*entity.DownloadResponse, error)
	}

	// VoiceService maneja la conexión y transmisión de audio a Discord
	VoiceService interface {
		// JoinChannel se une a un canal de voz
		JoinChannel(channelID string) error

		// LeaveChannel sale del canal de voz actual
		LeaveChannel() error

		// StreamAudio transmite audio al canal
		StreamAudio(ctx context.Context, audioData []byte) error
	}

	PlaylistService interface {
		// AddSong agrega una canción a la lista
		AddSong(song *entity.Song) error

		// GetNextSong obtiene la siguiente canción de la cola
		GetNextSong() (*entity.Song, error)

		// Clear limpia la lista de reproducción
		Clear() error
	}

	MessageService interface {
		// SendPlayingStatus envía el estado de reproducción actual
		SendPlayingStatus(channelID string, song *entity.Song) (string, error)
	}

	PlayerService interface {
		// Play inicia la reproducción de música
		Play(ctx context.Context, voiceChannelID string, textChannelID string) error

		// Stop detiene la reproducción
		Stop() error

		// Skip salta la canción actual
		Skip() error
	}
)
