package errors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrInvalidInput          = NewAppError("invalid_input", "Input inválido")
	ErrYouTubeAPIError       = NewAppError("youtube_api_error", "Error en la API de YouTube")
	ErrProviderNotFound      = NewAppError("provider_not_found", "Proveedor no encontrado")
	ErrStartOperationFailed  = NewAppError("start_operation_failed", "Error al iniciar la operación")
	ErrGetMediaDetailsFailed = NewAppError("get_media_details_failed", "Error al obtener detalles del media")

	ErrDownloadFailed            = NewAppError("download_failed", "Error en descarga de audio")
	ErrEncodingFailed            = NewAppError("encoding_failed", "Error en codificación de audio")
	ErrUploadFailed              = NewAppError("upload_failed", "Error al subir a almacenamiento")
	ErrDuplicateRecord           = NewAppError("duplicate_record", "El registro ya existe")
	ErrOperationInitFailed       = NewAppError("operation_init_failed", "Error al iniciar la operación")
	ErrOperationNotFound         = NewAppError("operation_not_found", "No se encontró la operación solicitada")
	ErrUpdateMediaFailed         = NewAppError("update_media_failed", "Error al actualizar el media")
	ErrPublishMessageFailed      = NewAppError("publish_message_failed", "Error al publicar el mensaje")
	ErrCodeSearchVideoIDFailed   = NewAppError("search_video_id_failed", "Error al buscar el ID del video")
	ErrCodeGetVideoDetailsFailed = NewAppError("get_video_details_failed", "Error al obtener detalles del video")

	ErrCodeDBConnectionFailed = NewAppError("db_connection_failed", "Error de conexión a la base de datos")
	ErrCodeInvalidVideoID     = NewAppError("invalid_video_id", "ID de video inválido")
	ErrCodeMediaNotFound      = NewAppError("media_not_found", "Media no encontrado")
	ErrCodeInvalidMetadata    = NewAppError("invalid_metadata", "Metadata inválido")
	ErrCodeSaveMediaFailed    = NewAppError("save_media_failed", "Error al guardar el media")
	ErrCodeDeleteMediaFailed  = NewAppError("delete_media_failed", "Error al eliminar el media")
	ErrCodeGetMediaFailed     = NewAppError("get_media_failed", "Error al obtener el media")
)

type AppError struct {
	Code    string
	Message string
	Err     error
}

func NewAppError(code, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

func (e *AppError) WithMessage(msg string) *AppError {
	return &AppError{
		Code:    e.Code,
		Message: msg,
		Err:     e.Err,
	}
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
	switch e.Code {
	case "invalid_input", "invalid_video_id", "invalid_metadata":
		return http.StatusBadRequest
	case "provider_not_found", "operation_not_found", "media_not_found":
		return http.StatusNotFound
	case "youtube_api_error":
		return http.StatusServiceUnavailable
	case "duplicate_record":
		return http.StatusConflict
	case "download_failed", "encoding_failed", "upload_failed",
		"operation_init_failed", "get_media_details_failed",
		"update_media_failed", "publish_message_failed", "start_operation_failed",
		"search_video_id_failed", "get_video_details_failed",
		"db_connection_failed", "save_media_failed", "delete_media_failed", "get_media_failed":
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func (e *AppError) Wrap(err error) *AppError {
	return &AppError{
		Code:    e.Code,
		Err:     err,
		Message: e.Message,
	}
}
