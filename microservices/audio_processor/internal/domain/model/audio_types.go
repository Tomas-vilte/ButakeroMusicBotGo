package model

import (
	"errors"
	"time"
)

var (
	ErrInvalidVolume           = errors.New("volumen fuera de los límites (0-512)")
	ErrInvalidFrameDuration    = errors.New("duración de fotograma inválida")
	ErrInvalidPacketLoss       = errors.New("porcentaje de pérdida de paquetes inválido")
	ErrInvalidAudioApplication = errors.New("aplicación de audio inválida")
	ErrInvalidCompressionLevel = errors.New("nivel de compresión fuera de los límites (0-10)")
	ErrInvalidThreads          = errors.New("la cantidad de hilos no puede ser negativa")
)

var (
	AudioApplicationVoip     AudioApplication = "voip"     // Aplicación optimizada para voz sobre IP (VoIP)
	AudioApplicationAudio    AudioApplication = "audio"    // Aplicación genérica para audio musical o de alta calidad
	AudioApplicationLowDelay AudioApplication = "lowdelay" // Aplicación optimizada para baja latencia

	// StdEncodeOptions define las opciones predeterminadas para la codificación de audio.
	StdEncodeOptions = &EncodeOptions{
		Volume:           256,                   // Volumen normal (256 = 100%)
		Channels:         2,                     // Audio estéreo (2 canales)
		FrameRate:        48000,                 // Frecuencia de muestreo estándar (48kHz)
		FrameDuration:    20,                    // Duración de frame típica para Opus (20ms)
		Bitrate:          96,                    // Bitrate de 96 kbps (calidad balanceada)
		Application:      AudioApplicationAudio, // Aplicación genérica por defecto
		CompressionLevel: 10,                    // Máxima compresión (mejor calidad)
		PacketLoss:       1,                     // 1% de pérdida de paquetes esperada
		BufferedFrames:   100,                   // Buffer para 100 frames (evita bloqueos)
		VBR:              true,                  // Usar bitrate variable (mejor calidad)
		StartTime:        0,                     // Inicio desde el segundo 0
	}
)

// AudioApplication representa el tipo de aplicación de audio para la codificación Opus.
// Define el perfil de codificación que afecta la calidad y latencia.
type AudioApplication string

// EncodeOptions contiene todas las configuraciones para la codificación de audio.
// Cada campo controla un aspecto específico del proceso de codificación.
type EncodeOptions struct {
	// Configuración básica
	Volume        int // Nivel de volumen (0-512, 256 = 100%)
	Channels      int // Número de canales (1 = mono, 2 = estéreo)
	FrameRate     int // Frecuencia de muestreo en Hz (ej: 48000)
	FrameDuration int // Duración del frame en ms (20, 40 o 60)
	Bitrate       int // Bitrate objetivo en kbps (8-512)
	PacketLoss    int // Porcentaje esperado de pérdida de paquetes (0-100)

	// Configuración avanzada
	RawOutput        bool             // Si true, omite metadatos en la salida
	Application      AudioApplication // Perfil de codificación (voip/audio/lowdelay)
	CoverFormat      string           // Formato de imagen para metadatos (ej: "jpeg")
	CompressionLevel int              // Nivel de compresión (0-10, 10 = mejor calidad)
	BufferedFrames   int              // Tamaño del buffer en frames
	VBR              bool             // Bitrate variable (true) o constante (false)
	Threads          int              // Número de hilos (0 = automático)
	StartTime        int              // Segundo de inicio para cortar el audio
	AudioFilter      string           // Filtros FFmpeg (ej: "volume=0.5")
}

// AudioFrame representa un fotograma de audio codificado.
type AudioFrame struct {
	Data     []byte // Datos binarios del frame
	Metadata bool   // Si true, contiene metadatos en lugar de audio
}

// EncodeStats contiene estadísticas sobre el proceso de codificación.
type EncodeStats struct {
	Size     int           // Tamaño del archivo codificado en KB
	Duration time.Duration // Duración total del audio procesado
	Bitrate  float32       // Bitrate promedio en kbps
	Speed    float32       // Velocidad de codificación (1.0 = tiempo real)
}

// AudioMetadata contiene metadatos técnicos sobre el audio codificado.
type AudioMetadata struct {
	Opus   *OpusMetadata   `json:"opus"`   // Metadatos específicos de Opus
	Origin *OriginMetadata `json:"origin"` // Metadatos sobre el origen del audio
}

// OriginMetadata describe el origen del archivo de audio.
type OriginMetadata struct {
	Source   string `json:"source"`   // Fuente (ej: "file", "stream")
	Bitrate  int    `json:"abr"`      // Bitrate original en bps
	Channels int    `json:"channels"` // Canales originales
	Encoding string `json:"encoding"` // Codec original (ej: "MP3")
	Url      string `json:"url"`      // URL de origen (si aplica)
}

// OpusMetadata contiene parámetros técnicos de la codificación Opus.
type OpusMetadata struct {
	Bitrate     int    `json:"abr"`         // Bitrate en bps
	SampleRate  int    `json:"sample_rate"` // Frecuencia de muestreo
	Application string `json:"mode"`        // Perfil usado (voip/audio/lowdelay)
	FrameSize   int    `json:"frame_size"`  // Tamaño del frame en muestras
	Channels    int    `json:"channels"`    // Número de canales
	VBR         bool   `json:"vbr"`         // Si usa bitrate variable
}

// PCMFrameLen calcula el tamaño de un frame PCM basado en la configuración.
// Retorna el número total de muestras PCM para un frame.
// Fórmula: 960 muestras/base * canales * (duración_frame / 20ms)
func (opts *EncodeOptions) PCMFrameLen() int {
	return 960 * opts.Channels * (opts.FrameDuration / 20)
}

// Validate verifica que todas las opciones de codificación estén dentro de rangos válidos.
// Retorna error si algún parámetro es inválido.
func (opts *EncodeOptions) Validate() error {
	if opts.Volume < 0 || opts.Volume > 512 {
		return ErrInvalidVolume
	}

	if opts.FrameDuration != 20 && opts.FrameDuration != 40 && opts.FrameDuration != 60 {
		return ErrInvalidFrameDuration
	}

	if opts.PacketLoss < 0 || opts.PacketLoss > 100 {
		return ErrInvalidPacketLoss
	}

	if opts.Application != AudioApplicationAudio &&
		opts.Application != AudioApplicationVoip &&
		opts.Application != AudioApplicationLowDelay {
		return ErrInvalidAudioApplication
	}

	if opts.CompressionLevel < 0 || opts.CompressionLevel > 10 {
		return ErrInvalidCompressionLevel
	}

	if opts.Threads < 0 {
		return ErrInvalidThreads
	}

	return nil
}
