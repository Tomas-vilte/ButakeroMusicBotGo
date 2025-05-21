package ports

import (
	"context"
	"io"
)

// Downloader es una interfaz que define el contrato para descargar audio.
type Downloader interface {
	DownloadAudio(ctx context.Context, url string) (io.Reader, error)
}
