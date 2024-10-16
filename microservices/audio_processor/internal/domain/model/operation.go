package model

// OperationResult representa el resultado de una operación de procesamiento de canción.
// Contiene información sobre la canción procesada, el estado de la operación,
// estadísticas de éxito y fallos, y datos relevantes sobre el archivo.
type (
	OperationResult struct {
		// PK es la clave primaria del resultado de la operación.
		// Se genera con UUID para garantizar la unicidad.
		PK string `json:"pk" dynamodbav:"PK"`

		SK string `json:"sk" dynamodbav:"SK"`

		// Status indica el estado actual de la operación.
		// Puede tomar valores como "In Progress", "Completed", "Failed", etc.
		Status string `json:"status"`

		// Message contiene un mensaje descriptivo sobre el resultado de la operación.
		// Puede incluir detalles de errores o mensajes de éxito.
		Message string `json:"message"`

		// Metadata contiene datos adicionales sobre la canción procesada.
		// Esto puede incluir información como el género, el artista, el álbum, etc.
		Metadata Metadata `json:"metadata"`

		// FileData contiene información sobre el archivo de la canción procesada.
		// Esto incluye la ruta del archivo, el tamaño del archivo, el tipo de archivo
		// y la URL pública del archivo.
		FileData FileData `json:"file_data"`

		// ProcessingDate registra la fecha en que se procesó la canción.
		// Se utiliza para registrar el momento de finalización del procesamiento.
		ProcessingDate string `json:"processing_date"`

		// Success indica si el procesamiento de la canción fue exitoso.
		// Es un valor booleano (true o false).
		Success bool `json:"success"`

		// Attempts registra la cantidad de intentos realizados para descargar o procesar la canción.
		// Se incrementa cada vez que se intenta realizar la operación.
		Attempts int `json:"attempts"`

		// Failures registra la cantidad de fallos ocurridos durante la descarga o el procesamiento.
		// Se utiliza para registrar y monitorear los problemas encontrados.
		Failures int `json:"failures"`
	}

	// FileData contiene información sobre el archivo de la canción procesada.
	// Esto incluye la ruta del archivo, el tamaño del archivo, el tipo de archivo
	// y la URL pública del archivo.
	FileData struct {
		// FilePath es la ruta del archivo de la canción procesada.
		FilePath string `json:"file_path"`

		// FileSize es el tamaño del archivo de la canción procesada.
		FileSize string `json:"file_size"`

		// FileType es el tipo de archivo de la canción procesada.
		FileType string `json:"file_type"`

		// PublicURL es la URL pública del archivo de la canción procesada.
		PublicURL string `json:"public_url"`
	}
)
