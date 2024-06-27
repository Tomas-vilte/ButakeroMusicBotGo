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
	maxOpusBlockSize = 8192 // Tamaño máximo del bloque de datos Opus
	maxOpusChunkSize = 4096 // Tamaño máximo de cada chunk de datos Opus
)

func NewDCAStreamerImpl(logger logging.Logger) *DCAStreamerImpl {
	return &DCAStreamerImpl{
		logger: logger,
	}
}

func (d *DCAStreamerImpl) StreamDCAData(ctx context.Context, dca io.Reader, opusChan chan<- []byte, positionCallback func(position time.Duration)) error {
	var opuslen int16
	framesSent := 0
	positionChan := make(chan int)
	opusBuf := make([]byte, maxOpusBlockSize)

	go func() {
		defer close(positionChan)
		for framesSent := range positionChan {
			positionCallback(time.Duration(framesSent) * frameLength)
		}
	}()

	for {
		err := binary.Read(dca, binary.LittleEndian, &opuslen)

		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			d.logger.Error("Error EOF o EOF inesperado encontrado durante la transmisión de datos DCA:", zap.Error(err))
			return nil
		}
		if err != nil {
			d.logger.Error("Error mientras se leia la longitud de DCA", zap.Error(err))
			return err
		}

		var bytesRead int
		var opusData []byte
		for bytesRead < int(opuslen) {
			n, err := dca.Read(opusBuf[:min(int(opuslen)-bytesRead, maxOpusBlockSize)])
			if err != nil {
				d.logger.Error("Error mientras se leia PCM de DCA:", zap.Error(err))
				return err
			}
			opusData = append(opusData, opusBuf[:n]...)
			bytesRead += n
		}

		for len(opusData) > maxOpusChunkSize {
			opusChan <- opusData[:maxOpusChunkSize]
			opusData = opusData[maxOpusChunkSize:]
		}
		if len(opusData) > 0 {
			opusChan <- opusData
		}

		framesSent++

		if positionCallback != nil && framesSent%50 == 0 {
			positionChan <- framesSent
		}

		select {
		case <-ctx.Done():
			return nil
		default:
		}
	}
}
