package model

// Metadata representa la información sobre una canción procesada.
// Contiene detalles sobre la canción, como su título, artista, duración, y URLs de recursos.
type Metadata struct {
	// ID es un identificador único para la canción.
	// Este campo se utiliza para asociar la metadata con una canción específica.
	ID string `json:"id"`

	// VideoID Es el identificador del video
	VideoID string `json:"video_id"`

	// Title es el título de la canción.
	// Representa el nombre de la canción tal como aparece en la fuente de origen.
	Title string `json:"title"`

	// Duration es la duración de la canción en segundos.
	// Representa cuánto tiempo dura la canción desde el inicio hasta el final.
	Duration string `json:"duration"`

	// URLS3 es la URL del archivo de la canción almacenado en Amazon S3.
	// Este campo se usa para acceder al archivo de audio descargado y procesado.
	URLS3 string `json:"url_s3"`

	// URLYouTube es la URL de la canción en YouTube.
	// Permite localizar la canción en YouTube para referencias adicionales o reproducción.
	URLYouTube string `json:"url_youtube"`

	// Thumbnail es la imagen en miniatura del contenido.
	// Proporciona una vista previa visual de la canción, útil para interfaces de usuario y presentación.
	Thumbnail string `json:"thumbnail"`

	// Platform indica la plataforma de origen de la canción (e.g., YouTube).
	// Este campo identifica la fuente desde la cual se obtuvo la canción.
	Platform string `json:"platform"`
}
