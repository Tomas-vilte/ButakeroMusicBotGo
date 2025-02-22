package ports

import (
	"context"
	"io"
)

type StorageAudio interface {
	// GetAudio obtiene el contenido del archivo de audio
	GetAudio(ctx context.Context, songPath string) (io.ReadCloser, error)
}
