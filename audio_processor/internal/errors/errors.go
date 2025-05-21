package errors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	errorStatusMap = map[string]int{
		"invalid_input":                http.StatusBadRequest,
		"invalid_video_id":             http.StatusBadRequest,
		"s3_invalid_file":              http.StatusBadRequest,
		"local_invalid_file":           http.StatusBadRequest,
		"provider_not_found":           http.StatusNotFound,
		"media_not_found":              http.StatusNotFound,
		"local_file_not_found":         http.StatusNotFound,
		"youtube_api_error":            http.StatusServiceUnavailable,
		"duplicate_record":             http.StatusConflict,
		"get_media_details_failed":     http.StatusInternalServerError,
		"update_media_failed":          http.StatusInternalServerError,
		"start_operation_failed":       http.StatusInternalServerError,
		"search_video_id_failed":       http.StatusInternalServerError,
		"get_video_details_failed":     http.StatusInternalServerError,
		"db_connection_failed":         http.StatusInternalServerError,
		"save_media_failed":            http.StatusInternalServerError,
		"delete_media_failed":          http.StatusInternalServerError,
		"search_songs_failed":          http.StatusInternalServerError,
		"get_media_failed":             http.StatusInternalServerError,
		"s3_upload_failed":             http.StatusInternalServerError,
		"s3_get_metadata_failed":       http.StatusInternalServerError,
		"s3_get_content_failed":        http.StatusInternalServerError,
		"local_upload_failed":          http.StatusInternalServerError,
		"local_get_metadata_failed":    http.StatusInternalServerError,
		"local_get_content_failed":     http.StatusInternalServerError,
		"local_directory_not_writable": http.StatusInternalServerError,
		"ytdlp_command_failed":         http.StatusInternalServerError,
		"ytdlp_invalid_output":         http.StatusInternalServerError,
		"kafka_connection_failed":      http.StatusInternalServerError,
		"kafka_topic_creation":         http.StatusInternalServerError,
		"kafka_publish_failed":         http.StatusInternalServerError,
		"kafka_consume_failed":         http.StatusInternalServerError,
		"kafka_tls_config":             http.StatusInternalServerError,
		"kafka_admin_error":            http.StatusInternalServerError,
		"mongodb_connection_failed":    http.StatusInternalServerError,
		"mongodb_tls_config":           http.StatusInternalServerError,
		"mongodb_ping_failed":          http.StatusInternalServerError,
		"mongodb_index_error":          http.StatusInternalServerError,
		"dynamodb_query_failed":        http.StatusInternalServerError,
		"dynamodb_marshal_failed":      http.StatusInternalServerError,
		"dynamodb_unmarshal_failed":    http.StatusInternalServerError,
		"sqs_publish_failed":           http.StatusInternalServerError,
		"sqs_consume_failed":           http.StatusInternalServerError,
		"sqs_delete_failed":            http.StatusInternalServerError,
		"sqs_deserialize_failed":       http.StatusInternalServerError,
	}
)

var (
	ErrInvalidInput          = NewAppError("invalid_input", "Input inválido")
	ErrYouTubeAPIError       = NewAppError("youtube_api_error", "Error en la API de YouTube")
	ErrProviderNotFound      = NewAppError("provider_not_found", "Proveedor no encontrado")
	ErrGetMediaDetailsFailed = NewAppError("get_media_details_failed", "Error al obtener detalles del media")

	ErrDuplicateRecord           = NewAppError("duplicate_record", "El registro ya existe")
	ErrUpdateMediaFailed         = NewAppError("update_media_failed", "Error al actualizar el media")
	ErrCodeSearchVideoIDFailed   = NewAppError("search_video_id_failed", "Error al buscar el ID del video")
	ErrCodeGetVideoDetailsFailed = NewAppError("get_video_details_failed", "Error al obtener detalles del video")

	ErrCodeDBConnectionFailed = NewAppError("db_connection_failed", "Error de conexión a la base de datos")
	ErrCodeInvalidVideoID     = NewAppError("invalid_video_id", "ID de video inválido")
	ErrCodeMediaNotFound      = NewAppError("media_not_found", "Media no encontrado")
	ErrCodeSaveMediaFailed    = NewAppError("save_media_failed", "Error al guardar el media")
	ErrCodeDeleteMediaFailed  = NewAppError("delete_media_failed", "Error al eliminar el media")
	ErrCodeSearchSongsFailed  = NewAppError("search_songs_failed", "Error al buscar canciones")

	ErrS3UploadFailed      = NewAppError("s3_upload_failed", "Error al subir archivo a S3")
	ErrS3GetMetadataFailed = NewAppError("s3_get_metadata_failed", "Error al obtener metadatos del archivo de S3")
	ErrS3GetContentFailed  = NewAppError("s3_get_content_failed", "Error al obtener contenido del archivo de S3")
	ErrS3InvalidFile       = NewAppError("s3_invalid_file", "El archivo proporcionado no es válido")

	ErrLocalUploadFailed         = NewAppError("local_upload_failed", "Error al subir archivo al almacenamiento local")
	ErrLocalGetMetadataFailed    = NewAppError("local_get_metadata_failed", "Error al obtener metadatos del archivo local")
	ErrLocalGetContentFailed     = NewAppError("local_get_content_failed", "Error al obtener contenido del archivo local")
	ErrLocalInvalidFile          = NewAppError("local_invalid_file", "El archivo proporcionado no es válido")
	ErrLocalFileNotFound         = NewAppError("local_file_not_found", "Archivo no encontrado en el almacenamiento local")
	ErrLocalDirectoryNotWritable = NewAppError("local_directory_not_writable", "El directorio no es escribible")

	ErrYTDLPCommandFailed = NewAppError("ytdlp_command_failed", "Error al ejecutar el comando yt-dlp")
	ErrYTDLPInvalidOutput = NewAppError("ytdlp_invalid_output", "Salida inválida de yt-dlp")

	ErrKafkaConnectionFailed = NewAppError("kafka_connection_failed", "Error de conexión con Kafka")
	ErrKafkaTopicCreation    = NewAppError("kafka_topic_creation", "Error al crear tópico")
	ErrKafkaMessagePublish   = NewAppError("kafka_publish_failed", "Error al publicar mensaje")
	ErrKafkaMessageConsume   = NewAppError("kafka_consume_failed", "Error al consumir mensaje")
	ErrKafkaTLSConfig        = NewAppError("kafka_tls_config", "Error en configuración TLS")
	ErrKafkaAdminClient      = NewAppError("kafka_admin_error", "Error en cliente administrativo")

	ErrSQSAWSConfig          = NewAppError("sqs_aws_config", "Error en configuración AWS")
	ErrSQSMessagePublish     = NewAppError("sqs_publish_failed", "Error al publicar mensaje")
	ErrSQSMessageConsume     = NewAppError("sqs_consume_failed", "Error al consumir mensaje")
	ErrSQSMessageDelete      = NewAppError("sqs_delete_failed", "Error al eliminar mensaje")
	ErrSQSMessageDeserialize = NewAppError("sqs_deserialize_failed", "Error al deserializar mensaje")

	ErrMongoDBConnectionFailed = NewAppError("mongodb_connection_failed", "Error de conexión con MongoDB")
	ErrMongoDBTLSConfig        = NewAppError("mongodb_tls_config", "Error en configuración TLS de MongoDB")
	ErrMongoDBPingFailed       = NewAppError("mongodb_ping_failed", "Error al verificar conexión con MongoDB")
	ErrMongoDBIndexCreation    = NewAppError("mongodb_index_error", "Error al crear índice en MongoDB")

	ErrDynamoDBQueryFailed     = NewAppError("dynamodb_query_failed", "Error al ejecutar query en DynamoDB")
	ErrDynamoDBMarshalFailed   = NewAppError("dynamodb_marshal_failed", "Error al serializar datos para DynamoDB")
	ErrDynamoDBUnmarshalFailed = NewAppError("dynamodb_unmarshal_failed", "Error al deserializar datos de DynamoDB")
)

type AppError struct {
	Code    string
	Message string
	Err     error
	VideoID string
}

func NewAppError(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func (e *AppError) WithMessage(msg ...string) *AppError {
	result := &AppError{
		Code: e.Code,
		Err:  e.Err,
	}

	if len(msg) > 0 {
		result.Message = msg[0]
		if len(msg) > 1 {
			result.VideoID = msg[1]
		}
	}
	return result
}

func IsAppError(err error) bool {
	var appError *AppError
	ok := errors.As(err, &appError)
	return ok
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) StatusCode() int {
	if status, ok := errorStatusMap[e.Code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

func (e *AppError) Wrap(err error) *AppError {
	return &AppError{
		Code:    e.Code,
		Err:     err,
		Message: e.Message,
	}
}
