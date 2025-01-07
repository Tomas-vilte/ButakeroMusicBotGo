package encoder

import (
	"bytes"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"io"
	"os"
	"sync"
	"time"
)

type (
	// AudioApplication representa el tipo de aplicación de audio.
	AudioApplication string

	// EncodeOptions contiene las opciones de configuración para la codificación de audio.
	EncodeOptions struct {
		Volume           int              // Nivel de volumen del audio (256 = normal)
		Channels         int              // Número de canales de audio
		FrameRate        int              // Frecuencia de muestreo del audio (por ej. 48000 Hz)
		FrameDuration    int              // Duración del marco de audio en ms (puede ser 20, 40 o 60 ms)
		Bitrate          int              // Tasa de bits en kb/s (puede ser de 8 a 128 kb/s)
		PacketLoss       int              // Porcentaje de pérdida de paquetes esperado
		RawOutput        bool             // Salida de opus en crudo (sin metadatos ni bytes mágicos)
		Application      AudioApplication // Aplicación de audio a utilizar
		CoverFormat      string           // Formato de la carátula (por ej. "jpeg")
		CompressionLevel int              // Nivel de compresión (0 - 10), donde 10 es mejor calidad pero más lenta codificación
		BufferedFrames   int              // Tamaño del búfer de cuadros
		VBR              bool             // Si se utiliza VBR (tasa de bits variable) o no
		Threads          int              // Número de hilos a utilizar (0 para automático)
		StartTime        int              // Tiempo de inicio de la secuencia de entrada en segundos
	}

	// Frame representa un marco de audio.
	Frame struct {
		data     []byte // Datos del marco de audio
		metadata bool   // Indica si el marco contiene metadatos
	}

	// EncodeStats contiene estadísticas sobre el proceso de codificación.
	EncodeStats struct {
		Size     int           // Tamaño del archivo codificado en bytes
		Duration time.Duration // Duración de la codificación
		Bitrate  float32       // Tasa de bits en kb/s
		Speed    float32       // Velocidad de procesamiento en fps (frames por segundo)
	}

	// encodeSessionImpl representa una sesión de codificación de audio.
	encodeSessionImpl struct {
		sync.Mutex                  // Mutex para sincronización concurrente
		options      *EncodeOptions // Opciones de codificación
		pipeReader   io.Reader      // Lector para el pipe
		filePath     string         // Ruta del archivo a codificar
		running      bool           // Indica si la sesión está en ejecución
		started      time.Time      // Hora de inicio de la sesión
		frameChannel chan *Frame    // Canal para transmitir los marcos de audio
		process      *os.Process    // Proceso de codificación
		lastStats    *EncodeStats   // Últimas estadísticas de codificación
		lastFrame    int            // Último marco procesado
		err          error          // Error que ocurrió durante la codificación
		ffmpegOutput string         // Salida del proceso ffmpeg
		logging      logger.Logger  // Logger para registros
		buf          bytes.Buffer   // Búfer para almacenar bytes no leídos (cuadros incompletos), utilizado para implementar io.Reader

	}

	Metadata struct {
		Opus   *OpusMetadata   `json:"opus"`
		Origin *OriginMetadata `json:"origin"`
	}

	OriginMetadata struct {
		Source   string `json:"source"`
		Bitrate  int    `json:"abr"`
		Channels int    `json:"channels"`
		Encoding string `json:"encoding"`
		Url      string `json:"url"`
	}

	OpusMetadata struct {
		Bitrate     int    `json:"abr"`
		SampleRate  int    `json:"sample_rate"`
		Application string `json:"mode"`
		FrameSize   int    `json:"frame_size"`
		Channels    int    `json:"channels"`
		VBR         bool   `json:"vbr"`
	}
)
