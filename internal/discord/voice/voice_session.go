package voice

import (
	"context"
	"io"
	"time"
)

// VoiceChatSession define métodos para interactuar con la sesión de voz del bot de Discord.
type VoiceChatSession interface {
	Close() error
	JoinVoiceChannel(channelID string) error
	LeaveVoiceChannel() error
	SendAudio(ctx context.Context, reader io.Reader, positionCallback func(time.Duration)) error
}

// PlayMessage es el mensaje que se enviará al canal de texto para mostrar la canción que se está reproduciendo actualmente.
type PlayMessage struct {
	Song     *Song
	Position time.Duration
}

// Song representa una canción que se puede reproducir.
type Song struct {
	Type          string
	Title         string
	URL           string
	Playable      bool
	ThumbnailURL  *string
	Duration      time.Duration
	StartPosition time.Duration
	RequestedBy   *string
}

// GetHumanName devuelve el nombre humano legible de la canción.
func (s *Song) GetHumanName() string {
	if s.Title != "" {
		return s.Title
	}
	return s.URL
}

// PlayedSong representa una canción que ha sido reproducida.
type PlayedSong struct {
	Song
	Position time.Duration
}
