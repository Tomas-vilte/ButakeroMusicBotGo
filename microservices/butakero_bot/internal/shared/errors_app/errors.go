package errors_app

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorCode string

const (
	// Errores generales
	ErrCodeInternalError   ErrorCode = "internal_error"
	ErrCodeInvalidInput    ErrorCode = "invalid_input"
	ErrCodeInvalidVideoID  ErrorCode = "invalid_video_id"
	ErrCodeInvalidMetadata ErrorCode = "invalid_metadata"

	// Errores relacionados con la API externa
	ErrCodeProviderNotFound      ErrorCode = "provider_not_found"
	ErrCodeYouTubeAPIError       ErrorCode = "youtube_api_error"
	ErrCodeAPIDuplicateRecord    ErrorCode = "duplicate_record"
	ErrCodeDownloadFailed        ErrorCode = "download_failed"
	ErrCodeEncodingFailed        ErrorCode = "encoding_failed"
	ErrCodeUploadFailed          ErrorCode = "upload_failed"
	ErrCodeOperationInitFailed   ErrorCode = "operation_init_failed"
	ErrCodeUpdateMediaFailed     ErrorCode = "update_media_failed"
	ErrCodePublishMessageFailed  ErrorCode = "publish_message_failed"
	ErrCodeOperationNotFound     ErrorCode = "operation_not_found"
	ErrCodeMediaNotFound         ErrorCode = "media_not_found"
	ErrCodeSearchVideoIDFailed   ErrorCode = "search_video_id_failed"
	ErrCodeGetVideoDetailsFailed ErrorCode = "get_video_details_failed"
	ErrCodeGetMediaDetailsFailed ErrorCode = "get_media_details_failed"

	// Errores relacionados con almacenamiento
	ErrCodeS3UploadFailed            ErrorCode = "s3_upload_failed"
	ErrCodeS3GetMetadataFailed       ErrorCode = "s3_get_metadata_failed"
	ErrCodeS3GetContentFailed        ErrorCode = "s3_get_content_failed"
	ErrCodeS3InvalidFile             ErrorCode = "s3_invalid_file"
	ErrCodeLocalUploadFailed         ErrorCode = "local_upload_failed"
	ErrCodeLocalGetMetadataFailed    ErrorCode = "local_get_metadata_failed"
	ErrCodeLocalGetContentFailed     ErrorCode = "local_get_content_failed"
	ErrCodeLocalInvalidFile          ErrorCode = "local_invalid_file"
	ErrCodeLocalFileNotFound         ErrorCode = "local_file_not_found"
	ErrCodeLocalDirectoryNotWritable ErrorCode = "local_directory_not_writable"
	ErrCodeSaveMediaFailed           ErrorCode = "save_media_failed"
	ErrCodeDeleteMediaFailed         ErrorCode = "delete_media_failed"
	ErrCodeGetMediaFailed            ErrorCode = "get_media_failed"

	// Errores relacionados con yt-dlp
	ErrCodeYTDLPCommandFailed ErrorCode = "ytdlp_command_failed"
	ErrCodeYTDLPInvalidOutput ErrorCode = "ytdlp_invalid_output"

	// Errores relacionados con Discord
	ErrCodeDiscordNotInVoiceChannel ErrorCode = "discord_not_in_voice_channel"
	ErrCodeDiscordFailedToAddSong   ErrorCode = "discord_failed_to_add_song"
	ErrCodeGuildPlayerNotFound      ErrorCode = "guild_player_not_found"

	// Errores relacionados con la base de datos
	ErrCodeDBConnectionFailed ErrorCode = "db_connection_failed"
	ErrCodeDBRecordNotFound   ErrorCode = "db_record_not_found"
)

var errorStatusMap = map[ErrorCode]int{
	// 400 Bad Request
	ErrCodeInvalidInput:     http.StatusBadRequest,
	ErrCodeInvalidVideoID:   http.StatusBadRequest,
	ErrCodeInvalidMetadata:  http.StatusBadRequest,
	ErrCodeS3InvalidFile:    http.StatusBadRequest,
	ErrCodeLocalInvalidFile: http.StatusBadRequest,

	// 404 Not Found
	ErrCodeProviderNotFound:  http.StatusNotFound,
	ErrCodeOperationNotFound: http.StatusNotFound,
	ErrCodeMediaNotFound:     http.StatusNotFound,
	ErrCodeLocalFileNotFound: http.StatusNotFound,
	ErrCodeDBRecordNotFound:  http.StatusNotFound,

	// 409 Conflict
	ErrCodeAPIDuplicateRecord: http.StatusConflict,

	// 503 Service Unavailable
	ErrCodeYouTubeAPIError: http.StatusServiceUnavailable,

	// 500 Internal Server Error
	ErrCodeInternalError:             http.StatusInternalServerError,
	ErrCodeDownloadFailed:            http.StatusInternalServerError,
	ErrCodeEncodingFailed:            http.StatusInternalServerError,
	ErrCodeUploadFailed:              http.StatusInternalServerError,
	ErrCodeOperationInitFailed:       http.StatusInternalServerError,
	ErrCodeUpdateMediaFailed:         http.StatusInternalServerError,
	ErrCodePublishMessageFailed:      http.StatusInternalServerError,
	ErrCodeSearchVideoIDFailed:       http.StatusInternalServerError,
	ErrCodeGetVideoDetailsFailed:     http.StatusInternalServerError,
	ErrCodeGetMediaDetailsFailed:     http.StatusInternalServerError,
	ErrCodeDBConnectionFailed:        http.StatusInternalServerError,
	ErrCodeSaveMediaFailed:           http.StatusInternalServerError,
	ErrCodeDeleteMediaFailed:         http.StatusInternalServerError,
	ErrCodeGetMediaFailed:            http.StatusInternalServerError,
	ErrCodeS3UploadFailed:            http.StatusInternalServerError,
	ErrCodeS3GetMetadataFailed:       http.StatusInternalServerError,
	ErrCodeS3GetContentFailed:        http.StatusInternalServerError,
	ErrCodeLocalUploadFailed:         http.StatusInternalServerError,
	ErrCodeLocalGetMetadataFailed:    http.StatusInternalServerError,
	ErrCodeLocalGetContentFailed:     http.StatusInternalServerError,
	ErrCodeLocalDirectoryNotWritable: http.StatusInternalServerError,
	ErrCodeYTDLPCommandFailed:        http.StatusInternalServerError,
	ErrCodeYTDLPInvalidOutput:        http.StatusInternalServerError,
	ErrCodeDiscordNotInVoiceChannel:  http.StatusInternalServerError,
	ErrCodeDiscordFailedToAddSong:    http.StatusInternalServerError,
	ErrCodeGuildPlayerNotFound:       http.StatusInternalServerError,
}

type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func NewAppError(code ErrorCode, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func IsAppError(err error) bool {
	var appError *AppError
	ok := errors.As(err, &appError)
	return ok
}

// StatusCode devuelve el código de estado HTTP correspondiente al código de error.
func (e *AppError) StatusCode() int {
	if status, ok := errorStatusMap[e.Code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// apiErrorMapping es un mapa que asocia los códigos de error de la API externa con los errores personalizados.
var apiErrorMapping = map[string]*AppError{
	"duplicate_record":       NewAppError(ErrCodeAPIDuplicateRecord, "La canción con ID '%s' ya está registrada", nil),
	"provider_not_found":     NewAppError(ErrCodeProviderNotFound, "El proveedor de música no fue encontrado", nil),
	"youtube_api_error":      NewAppError(ErrCodeYouTubeAPIError, "Error en la API de YouTube", nil),
	"download_failed":        NewAppError(ErrCodeDownloadFailed, "Error al descargar la canción", nil),
	"encoding_failed":        NewAppError(ErrCodeEncodingFailed, "Error al codificar la canción", nil),
	"upload_failed":          NewAppError(ErrCodeUploadFailed, "Error al subir la canción", nil),
	"operation_init_failed":  NewAppError(ErrCodeOperationInitFailed, "Error al iniciar la operación", nil),
	"update_media_failed":    NewAppError(ErrCodeUpdateMediaFailed, "Error al actualizar los medios", nil),
	"publish_message_failed": NewAppError(ErrCodePublishMessageFailed, "Error al publicar el mensaje", nil),

	"invalid_input":            NewAppError(ErrCodeInvalidInput, "Entrada inválida", nil),
	"invalid_video_id":         NewAppError(ErrCodeInvalidVideoID, "ID de video inválido", nil),
	"invalid_metadata":         NewAppError(ErrCodeInvalidMetadata, "Metadata inválida", nil),
	"media_not_found":          NewAppError(ErrCodeMediaNotFound, "Media no encontrado", nil),
	"operation_not_found":      NewAppError(ErrCodeOperationNotFound, "Operación no encontrada", nil),
	"get_media_details_failed": NewAppError(ErrCodeGetMediaDetailsFailed, "Error al obtener detalles del media", nil),
	"search_video_id_failed":   NewAppError(ErrCodeSearchVideoIDFailed, "Error al buscar el ID del video", nil),
	"get_video_details_failed": NewAppError(ErrCodeGetVideoDetailsFailed, "Error al obtener detalles del video", nil),
	"db_connection_failed":     NewAppError(ErrCodeDBConnectionFailed, "Error de conexión a la base de datos", nil),
	"save_media_failed":        NewAppError(ErrCodeSaveMediaFailed, "Error al guardar el media", nil),
	"delete_media_failed":      NewAppError(ErrCodeDeleteMediaFailed, "Error al eliminar el media", nil),
	"get_media_failed":         NewAppError(ErrCodeGetMediaFailed, "Error al obtener el media", nil),

	"s3_upload_failed":             NewAppError(ErrCodeS3UploadFailed, "Error al subir archivo a S3", nil),
	"s3_get_metadata_failed":       NewAppError(ErrCodeS3GetMetadataFailed, "Error al obtener metadatos del archivo de S3", nil),
	"s3_get_content_failed":        NewAppError(ErrCodeS3GetContentFailed, "Error al obtener contenido del archivo de S3", nil),
	"s3_invalid_file":              NewAppError(ErrCodeS3InvalidFile, "El archivo proporcionado no es válido", nil),
	"local_upload_failed":          NewAppError(ErrCodeLocalUploadFailed, "Error al subir archivo al almacenamiento local", nil),
	"local_get_metadata_failed":    NewAppError(ErrCodeLocalGetMetadataFailed, "Error al obtener metadatos del archivo local", nil),
	"local_get_content_failed":     NewAppError(ErrCodeLocalGetContentFailed, "Error al obtener contenido del archivo local", nil),
	"local_invalid_file":           NewAppError(ErrCodeLocalInvalidFile, "El archivo proporcionado no es válido", nil),
	"local_file_not_found":         NewAppError(ErrCodeLocalFileNotFound, "Archivo no encontrado en el almacenamiento local", nil),
	"local_directory_not_writable": NewAppError(ErrCodeLocalDirectoryNotWritable, "El directorio no es escribible", nil),

	"ytdlp_command_failed": NewAppError(ErrCodeYTDLPCommandFailed, "Error al ejecutar el comando yt-dlp", nil),
	"ytdlp_invalid_output": NewAppError(ErrCodeYTDLPInvalidOutput, "Salida inválida de yt-dlp", nil),
}

// GetAPIError devuelve un AppError basado en el código de error de la API externa.
func GetAPIError(code string) *AppError {
	if appErr, ok := apiErrorMapping[code]; ok {
		return appErr
	}
	return NewAppError(ErrCodeInternalError, "Error desconocido en la API", nil)
}
