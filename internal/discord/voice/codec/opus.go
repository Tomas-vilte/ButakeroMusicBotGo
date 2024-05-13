package codec

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"time"
)

type DCAStreamer interface {
	StreamDCAData(ctx context.Context, dca io.Reader, opusChan chan<- []byte, positionCallback func(position time.Duration)) error
}

type DCAStreamerImpl struct{}

const (
	frameLength = time.Duration(20) * time.Millisecond
)

func (d *DCAStreamerImpl) StreamDCAData(ctx context.Context, dca io.Reader, opusChan chan<- []byte, positionCallback func(position time.Duration)) error {
	var opuslen int16
	framesSent := 0

	for {
		err := binary.Read(dca, binary.LittleEndian, &opuslen)

		if err == io.EOF || errors.Is(err, io.ErrUnexpectedEOF) {
			log.Printf("Error EOF o inesperado EOF encontrado en la transmisión de datos DCA: %v\n", err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("while reading length from DCA: %w", err)
		}

		inBuf := make([]byte, opuslen)
		err = binary.Read(dca, binary.LittleEndian, &inBuf)

		if err != nil {
			return fmt.Errorf("while reading PCM from DCA: %w", err)
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
