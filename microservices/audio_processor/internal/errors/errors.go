package errors

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrInvalidInput     = NewAppError("invalid_input", "Input inválido")
	ErrProviderNotFound = NewAppError("provider_not_found", "Proveedor no encontrado")
	ErrExternalAPIError = NewAppError("external_api_error", "Error en API externa")

	ErrDownloadFailed      = NewAppError("download_failed", "Error en descarga de audio")
	ErrEncodingFailed      = NewAppError("encoding_failed", "Error en codificación de audio")
	ErrUploadFailed        = NewAppError("upload_failed", "Error al subir a almacenamiento")
	ErrMetadataSaveFailed  = NewAppError("metadata_failed", "Error al guardar metadatos")
	ErrMessageSendFailed   = NewAppError("message_failed", "Error al enviar mensaje")
	ErrOperationInitFailed = NewAppError("operation_init_failed", "Error al iniciar la operación")
	ErrOperationNotFound   = NewAppError("operation_not_found", "No se encontró la operación solicitada")
	ErrInvalidUUID         = NewAppError("invalid_uuid", "UUID inválido")
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

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) StatusCode() int {
	switch e.Code {
	case "invalid_input", "invalid_uuid":
		return http.StatusBadRequest
	case "provider_not_found", "media_not_found", "operation_not_found":
		return http.StatusNotFound
	case "external_api_error":
		return http.StatusServiceUnavailable
	case "download_failed", "encoding_failed", "upload_failed",
		"metadata_failed", "message_failed", "empty_buffer",
		"operation_init_failed":
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

func (e *AppError) Is(target error) bool {
	var t *AppError
	ok := errors.As(target, &t)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

func (e *AppError) Unwrap() error {
	return e.Err
}
