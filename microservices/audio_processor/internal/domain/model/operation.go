package model

import "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

// OperationResult representa el resultado de una operación de procesamiento de canción.
// Contiene información sobre la canción procesada, el estado de la operación,
// estadísticas de éxito y fallos, y datos relevantes sobre el archivo.
type (
	OperationResult struct {
		// ID es la clave primaria del resultado de la operación.
		// Se genera con UUID para garantizar la unicidad.
		ID string `bson:"_id" json:"id" dynamodbav:"PK"`

		SK string `bson:"sk" json:"sk" dynamodbav:"SK"`

		// Status indica el estado actual de la operación.
		// Puede tomar valores como "In Progress", "Completed", "Failed", etc.
		Status string `bson:"status" json:"status" dynamodbav:"status"`

		// Message contiene un mensaje descriptivo sobre el resultado de la operación.
		// Puede incluir detalles de errores o mensajes de éxito.
		Message string `bson:"message" json:"message" dynamodbav:"Message"`

		// Metadata contiene datos adicionales sobre la canción procesada.
		// Esto puede incluir información como el género, el artista, el álbum, etc.
		Metadata *Metadata `bson:"metadata" json:"metadata" dynamodbav:"metadata"`

		// FileData contiene información sobre el archivo de la canción procesada.
		// Esto incluye la ruta del archivo, el tamaño del archivo, el tipo de archivo
		// y la URL pública del archivo.
		FileData *FileData `bson:"file_data" json:"file_data" dynamodbav:"file_data"`

		// ProcessingDate registra la fecha en que se procesó la canción.
		// Se utiliza para registrar el momento de finalización del procesamiento.
		ProcessingDate string `bson:"processing_date" json:"processing_date" dynamodbav:"processing_date"`

		// Success indica si el procesamiento de la canción fue exitoso.
		// Es un valor booleano (true o false).
		Success bool `bson:"success" json:"success" dynamodbav:"success"`

		// Attempts registra la cantidad de intentos realizados para descargar o procesar la canción.
		// Se incrementa cada vez que se intenta realizar la operación.
		Attempts int `bson:"attempts" json:"attempts" dynamodbav:"attempts"`

		// Failures registra la cantidad de fallos ocurridos durante la descarga o el procesamiento.
		// Se utiliza para registrar y monitorear los problemas encontrados.
		Failures int `bson:"failures" json:"failures" dynamodbav:"failures"`
	}

	// FileData contiene información sobre el archivo de la canción procesada.
	// Esto incluye la ruta del archivo, el tamaño del archivo, el tipo de archivo
	// y la URL pública del archivo.
	FileData struct {
		// FilePath es la ruta del archivo de la canción procesada.
		FilePath string `bson:"file_path" json:"file_path" dynamodbav:"file_path"`

		// FileSize es el tamaño del archivo de la canción procesada.
		FileSize string `bson:"file_size" json:"file_size" dynamodbav:"file_size"`

		// FileType es el tipo de archivo de la canción procesada.
		FileType string `bson:"file_type" json:"file_type" dynamodbav:"file_type"`
	}

	OperationInitResult struct {
		ID        string `json:"operation_id"`
		SongID    string `json:"song_id"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
	}
)

func (f *FileData) ToAttributeValue() map[string]types.AttributeValue {
	return map[string]types.AttributeValue{
		"file_path": &types.AttributeValueMemberS{Value: f.FilePath},
		"file_size": &types.AttributeValueMemberS{Value: f.FileSize},
		"file_type": &types.AttributeValueMemberS{Value: f.FileType},
	}
}
