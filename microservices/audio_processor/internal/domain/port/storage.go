package port

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"io"
)

type (
	// Storage define la interfaz para interactuar con un servicio de almacenamiento.
	// Permite subir archivos y obtener metadatos de archivos.
	Storage interface {
		// UploadFile sube un archivo al servicio de almacenamiento con la clave especificada.
		UploadFile(ctx context.Context, key string, body io.Reader) error

		// GetFileMetadata obtiene los metadatos del archivo con la clave especificada.
		GetFileMetadata(ctx context.Context, key string) (*model.FileData, error)
	}
)
