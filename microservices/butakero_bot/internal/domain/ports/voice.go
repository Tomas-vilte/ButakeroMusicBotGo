package ports

import (
	"context"
)

type VoiceSession interface {
	Close() error
	JoinVoiceChannel(channelID string) error
	LeaveVoiceChannel() error
	SendAudio(ctx context.Context, frames []byte) error
}
