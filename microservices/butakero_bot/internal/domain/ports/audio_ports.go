package ports

import (
	"context"
	"io"
)

type (
	// Decoder define una interfaz para decodificar audio descargado a un formato que Discord entiende.
	Decoder interface {
		// OpusFrame devuelve un marco de audio en formato Opus.
		OpusFrame() ([]byte, error)
		// Close cierra el decodificador.
		Close() error
	}

	// StorageAudio define métodos para obtener el contenido de archivos de audio.
	StorageAudio interface {
		// GetAudio obtiene el contenido del archivo de audio
		GetAudio(ctx context.Context, songPath string) (io.ReadCloser, error)
	}
)
