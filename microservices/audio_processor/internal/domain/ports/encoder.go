package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/encoder"
	"io"
)

type AudioEncoder interface {
	Encode(ctx context.Context, r io.Reader, options *encoder.EncodeOptions) (encoder.EncodeSession, error)
}
