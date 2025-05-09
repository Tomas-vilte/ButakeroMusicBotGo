package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"io"
)

type (
	AudioEncoder interface {
		Encode(ctx context.Context, r io.Reader, options *model.EncodeOptions) (EncodeSession, error)
	}

	EncodeSession interface {
		ReadFrame() ([]byte, error)
		Read(p []byte) (n int, err error)
		Stop() error
		FFMPEGMessages() string
		Cleanup()
	}
)
