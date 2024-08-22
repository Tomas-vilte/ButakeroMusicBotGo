package decoder

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"
	"time"

	"github.com/Tomas-vilte/GoMusicBot/internal/types"
)

// Decoder es una estructura que proporciona métodos para leer y procesar datos de un flujo de entrada.
type Decoder struct {
	r                   io.Reader       // Reader que lee datos del flujo de entrada
	Metadata            *types.Metadata // Metadatos leídos del flujo
	firstFrameProcessed bool            // Indica si el primer marco ha sido procesado
}

// NewDecoder crea una nueva instancia de Decoder, inicializando el Reader con un bufio.Reader para un buffering eficiente.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: bufio.NewReader(r),
	}
}

// ReadMetadata lee y procesa los metadatos del flujo de entrada. Este método debe ser llamado antes de leer cualquier marco de datos.
func (d *Decoder) ReadMetadata() error {
	if d.firstFrameProcessed {
		return ErrNotFirstFrame // Retorna un error si el primer marco ya ha sido procesado
	}
	d.firstFrameProcessed = true

	// Lee los primeros 4 bytes del flujo, que se utilizan para determinar el tamaño de los metadatos
	header := make([]byte, 4)
	if _, err := io.ReadFull(d.r, header); err != nil {
		return err // Retorna un error si no se pudieron leer los 4 bytes
	}

	// Lee el tamaño de los metadatos
	var metaLen int32
	if err := binary.Read(d.r, binary.LittleEndian, &metaLen); err != nil {
		return err // Retorna un error si no se pudo leer el tamaño de los metadatos
	}

	// Lee los metadatos basados en el tamaño leído
	jsonBuf := make([]byte, metaLen)
	if _, err := io.ReadFull(d.r, jsonBuf); err != nil {
		return err // Retorna un error si no se pudieron leer los metadatos
	}

	// Deserializa los metadatos en la estructura Metadata
	d.Metadata = new(types.Metadata)
	return json.Unmarshal(jsonBuf, d.Metadata)
}

// OpusFrame lee y decodifica un marco Opus del flujo de entrada. Si es la primera vez que se llama, también lee los metadatos.
func (d *Decoder) OpusFrame() ([]byte, error) {
	if !d.firstFrameProcessed {
		if err := d.ReadMetadata(); err != nil {
			return nil, err // Retorna un error si no se pudieron leer los metadatos
		}
	}
	return DecodeFrame(d.r) // Lee y decodifica un marco Opus
}

// FrameDuration devuelve la duración del marco en función de los metadatos. Si no hay metadatos, devuelve una duración predeterminada de 20 ms.
func (d *Decoder) FrameDuration() time.Duration {
	if d.Metadata == nil {
		return 20 * time.Millisecond // Valor predeterminado si no hay metadatos
	}
	return time.Duration(((d.Metadata.Opus.FrameSize/d.Metadata.Opus.Channels)/960)*20) * time.Millisecond
}

// DecodeFrame lee y decodifica un marco del flujo de entrada. El tamaño del marco se lee como un entero de 16 bits.
func DecodeFrame(r io.Reader) ([]byte, error) {
	var size int16
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return nil, err // Retorna un error si no se pudo leer el tamaño del marco
	}

	if size < 0 {
		return nil, ErrNegativeFrameSize // Retorna un error si el tamaño del marco es negativo
	}

	frame := make([]byte, size)
	if _, err := io.ReadFull(r, frame); err != nil {
		return nil, err // Retorna un error si no se pudo leer el marco
	}

	return frame, nil
}
