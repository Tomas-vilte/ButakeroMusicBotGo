package codec

import (
	"context"
	"encoding/binary"
	"errors"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"go.uber.org/zap"
	"io"
	"time"
)

type DCAStreamer interface {
	StreamDCAData(ctx context.Context, dca io.Reader, opusChan chan<- []byte, positionCallback func(position time.Duration)) error
}

type DCAStreamerImpl struct {
	logger logging.Logger
}

const (
	frameLength      = time.Duration(20) * time.Millisecond
	maxOpusBlockSize = 16384 // Tamaño máximo del bloque de datos Opus
)

func NewDCAStreamerImpl(logger logging.Logger) *DCAStreamerImpl {
	return &DCAStreamerImpl{
		logger: logger,
	}
}

func (d *DCAStreamerImpl) StreamDCAData(ctx context.Context, dca io.Reader, opusChan chan<- []byte, positionCallback func(position time.Duration)) error {
	// Declaración de variables
	var opuslen int16
	framesSent := 0
	positionChan := make(chan int)
	opusBuf := make([]byte, maxOpusBlockSize)

	// Goroutine para la función de callback de posición
	go func() {
		defer close(positionChan) // Cerramos el canal al finalizar la goroutine
		for framesSent := range positionChan {
			positionCallback(time.Duration(framesSent) * frameLength)
		}
	}()

	// Bucle infinito para la transmisión de datos DCA
	for {
		// Lectura de la longitud del bloque de datos Opus
		err := binary.Read(dca, binary.LittleEndian, &opuslen)

		// Manejo de errores al leer la longitud de los datos Opus
		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			d.logger.Error("Error EOF o EOF inesperado encontrado durante la transmisión de datos DCA:", zap.Error(err))
			return nil
		}
		if err != nil {
			d.logger.Error("Error mientras se leia la longitud de DCA", zap.Error(err))
			return err
		}

		// Lectura de los datos Opus en bloques más grandes
		var bytesRead int
		for bytesRead < int(opuslen) {
			n, err := dca.Read(opusBuf[:min(int(opuslen)-bytesRead, maxOpusBlockSize)])
			if err != nil {
				d.logger.Error("Error mientras se leia PCM de DCA:", zap.Error(err))
				return err
			}
			opusChunk := make([]byte, n)
			copy(opusChunk, opusBuf[:n])
			opusChan <- opusChunk
			bytesRead += n
		}

		// Incremento de frames enviados
		framesSent++

		// Llamada a la función de callback de posición si es necesario
		if positionCallback != nil && framesSent%50 == 0 {
			positionChan <- framesSent
		}

		// Verificar si el contexto ha sido cancelado
		select {
		case <-ctx.Done():
			return nil
		default:
		}
	}
}
