package ports

import (
	"context"
	"io"
)

// StorageAudio define métodos para obtener el contenido de archivos de audio.
type StorageAudio interface {
	// GetAudio obtiene el contenido del archivo de audio
	GetAudio(ctx context.Context, songPath string) (io.ReadCloser, error)
}
