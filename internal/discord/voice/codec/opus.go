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
	frameLength = time.Duration(20) * time.Millisecond
)

func NewDCAStreamerImpl(logger logging.Logger) *DCAStreamerImpl {
	return &DCAStreamerImpl{
		logger: logger,
	}
}

func (d *DCAStreamerImpl) StreamDCAData(ctx context.Context, dca io.Reader, opusChan chan<- []byte, positionCallback func(position time.Duration)) error {
	var opuslen int16
	framesSent := 0

	for {
		err := binary.Read(dca, binary.LittleEndian, &opuslen)

		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			d.logger.Error("Error EOF o EOF inesperado encontrado durante la transmisión de datos DCA:", zap.Error(err))
			//log.Printf("Error: EOF o EOF inesperado encontrado durante la transmisión de datos DCA: %v\n", err)
			return nil
		}

		if err != nil {
			d.logger.Error("Error mientras se leia la longitud de DCA", zap.Error(err))
			return err
		}

		inBuf := make([]byte, opuslen)
		err = binary.Read(dca, binary.LittleEndian, &inBuf)

		if err != nil {
			d.logger.Error("Error mientras se leia PCM de DCA:", zap.Error(err))
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case opusChan <- inBuf:
			framesSent++
			if positionCallback != nil && framesSent%50 == 0 {
				go func() {
					positionCallback(time.Duration(framesSent) * frameLength)
				}()
			}
		}
	}
}
