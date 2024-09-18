package encoder

import "errors"

var (
	ErrInvalidVolume           = errors.New("volumen fuera de los límites (0-512)")
	ErrInvalidFrameDuration    = errors.New("duración de fotograma inválida")
	ErrInvalidPacketLoss       = errors.New("porcentaje de pérdida de paquetes inválido")
	ErrInvalidAudioApplication = errors.New("aplicación de audio inválida")
	ErrInvalidCompressionLevel = errors.New("nivel de compresión fuera de los límites (0-10)")
	ErrInvalidThreads          = errors.New("la cantidad de hilos no puede ser negativa")
	ErrFailedToStartFFMPEG     = errors.New("error al iniciar ffmpeg")
	ErrFailedToReadStdout      = errors.New("error al leer la salida estándar de ffmpeg")
	ErrFailedToReadStderr      = errors.New("error al leer la salida de error estándar de ffmpeg")
)