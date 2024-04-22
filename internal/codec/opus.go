package codec

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"time"
)

const (
	frameLength = time.Duration(20) * time.Millisecond
)

// Decoder es una interfaz que define los metodos para decodiciar a dca
type Decoder interface {
	Decode() ([]byte, error)
}

type FileDecoder struct {
	reader io.Reader
}

func NewFileDecoder(reader io.Reader) *FileDecoder {
	return &FileDecoder{
		reader: reader,
	}
}

func (fd *FileDecoder) Decode() ([]byte, error) {
	var opuslen int16
	err := binary.Read(fd.reader, binary.LittleEndian, &opuslen)
	if err != nil {
		return nil, fmt.Errorf("error al leer la longitud desde el archivo DCA: %w", err)
	}

	inBuf := make([]byte, opuslen)
	err = binary.Read(fd.reader, binary.LittleEndian, &inBuf)
	if err != nil {
		return nil, fmt.Errorf("error al leer PCM desde el archivo DCA: %w", err)
	}
	return inBuf, nil
}

func StreamDCAData(ctx context.Context, decoder Decoder, opusChan chan<- []byte, positionCallback func(position time.Duration)) error {
	framesSent := 0

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			inBuf, err := decoder.Decode()
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					log.Println("Fin del archivo DCA alcanzado.")
					return nil
				}
				log.Printf("Error al decodificar datos DCA: %v\n", err)
				return err
			}

			opusChan <- inBuf
			framesSent++

			if positionCallback != nil && framesSent%50 == 0 {
				positionCallback(time.Duration(framesSent) * frameLength)
			}
		}
	}
}
