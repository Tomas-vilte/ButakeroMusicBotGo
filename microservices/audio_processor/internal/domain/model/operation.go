package model

// OperationResult representa el resultado de una operación de procesamiento de canción.
// Contiene información sobre la canción procesada, el estado de la operación,
// y estadísticas de éxito y fallos.
type OperationResult struct {
	// ID es un identificador único para el resultado de la operación.
	// Se genera utilizando UUID para asegurar que cada resultado tenga una clave única.
	ID string `json:"id"`

	// SongID es el identificador único de la canción que se está procesando.
	// Este ID se utiliza para asociar el resultado de la operación con una canción específica.
	SongID string `json:"song_id"`

	// Status indica el estado actual de la operación.
	// Puede tener valores como "In Progress", "Completed", "Failed", etc.
	Status string `json:"status"`

	// Message contiene un mensaje descriptivo relacionado con el resultado de la operación.
	// Puede incluir detalles de errores o mensajes de éxito.
	Message string `json:"message"`

	// Data es un campo opcional que puede contener información adicional sobre la operación.
	// Generalmente se utiliza para almacenar datos relacionados con el procesamiento de la canción,
	// como el tamaño del archivo, duración, etc.
	Data string `json:"data"`

	// ProcessingDate es la fecha en la que se procesó la canción.
	// Se utiliza para registrar cuándo se completó el procesamiento.
	ProcessingDate string `json:"processing_date"`

	// Success indica si el procesamiento de la canción fue exitoso o no.
	// Es un campo booleano que puede ser true o false.
	Success bool `json:"success"`

	// Attempts es el número de intentos realizados para descargar o procesar la canción.
	// Se incrementa cada vez que se intenta realizar la operación.
	Attempts int `json:"attempts"`

	// Failures es el número de fallos ocurridos durante el proceso de descarga o procesamiento.
	// Se utiliza para registrar y monitorizar los problemas encontrados.
	Failures int `json:"failures"`
}
