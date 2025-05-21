package ports

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"io"
)

type (
	// Storage define la interfaz para interactuar con un servicio de almacenamiento.
	// Permite subir archivos, obtener metadatos de archivos y obtener el contenido de archivos.
	Storage interface {
		// UploadFile sube un archivo al servicio de almacenamiento con la clave especificada.
		UploadFile(ctx context.Context, key string, body io.Reader) error

		// GetFileMetadata obtiene los metadatos del archivo con la clave especificada.
		GetFileMetadata(ctx context.Context, key string) (*model.FileData, error)

		// GetFileContent obtiene el contenido del archivo con la clave especificada.
		GetFileContent(ctx context.Context, path string, key string) (io.ReadCloser, error)
	}
)
