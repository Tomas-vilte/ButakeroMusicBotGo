package model

type Metadata struct {
	ID             string `json:"id"`              // Identificador único de la canción.
	Title          string `json:"title"`           // Título de la canción.
	Artist         string `json:"artist"`          // Artista de la canción.
	Duration       int    `json:"duration"`        // Duración de la canción en segundos.
	URLS3          string `json:"url_s3"`          // URL del archivo almacenado en S3.
	URLYouTube     string `json:"url_youtube"`     // URL de la canción en YouTube.
	Thumbnail      string `json:"thumbnail"`       // Imagen del contenido
	DownloadDate   string `json:"download_date"`   // Fecha en que se descargó la canción.
	Platform       string `json:"platform"`        // Plataforma de origen (e.g., YouTube).
	ProcessingDate string `json:"processing_date"` // Fecha en que se procesó la canción.
	Success        bool   `json:"success"`         // Indica si el procesamiento fue exitoso.
	Attempts       int    `json:"attempts"`        // Número de intentos para descargar o procesar la canción.
	Failures       int    `json:"failures"`        // Número de fallos durante el proceso.
}
