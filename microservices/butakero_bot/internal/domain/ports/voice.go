package ports

import (
	"context"
	"io"
)

// VoiceSession define una interfaz para manejar sesiones de voz.
type VoiceSession interface {
	// JoinVoiceChannel une a un canal de voz especificado por channelID.
	JoinVoiceChannel(channelID string) error
	// LeaveVoiceChannel deja el canal de voz actual.
	LeaveVoiceChannel() error
	// SendAudio envía audio a través de la sesión de voz.
	SendAudio(ctx context.Context, reader io.ReadCloser) error
}
