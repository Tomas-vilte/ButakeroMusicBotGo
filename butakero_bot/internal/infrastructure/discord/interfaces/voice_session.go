package interfaces

import (
	"context"
)

// VoiceConnection define una interfaz para manejar sesiones de voz.
type VoiceConnection interface {
	// JoinVoiceChannel une a un canal de voz especificado por channelID.
	JoinVoiceChannel(ctx context.Context, channelID string) error
	// LeaveVoiceChannel deja el canal de voz actual.
	LeaveVoiceChannel(ctx context.Context) error
	// SendAudio envía audio a través de la sesión de voz.
	SendAudio(ctx context.Context, audioDecoder Decoder) error
	// Pause pausa la sesión de voz.
	Pause()
	// Resume reanuda la sesión de voz.
	Resume()
	// IsConnected verifica si esta conectado al chat de voz
	IsConnected() bool
}
