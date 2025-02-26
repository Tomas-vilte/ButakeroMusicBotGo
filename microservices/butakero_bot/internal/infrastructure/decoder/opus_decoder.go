package decoder

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"time"
)

var (
	ErrNotFirstFrame     = errors.New("la metadata solo puede encontrarse en el primer marco")
	ErrNegativeFrameSize = errors.New("tamaño del marco es negativo, posiblemente está corrupto")
	ErrInvalidMetadata   = errors.New("metadata inválida")
	ErrDecoderClosed     = errors.New("el decodificador está cerrado")
)

type OpusDecoder struct {
	reader              io.Reader
	closer              io.Closer
	metadata            *Metadata
	firstFrameProcessed bool
	closed              bool
}

func NewOpusDecoder(r io.ReadCloser) *OpusDecoder {
	return &OpusDecoder{
		reader: r,
		closed: false,
	}
}

func NewBufferedOpusDecoder(r io.ReadCloser) *OpusDecoder {
	return &OpusDecoder{
		reader:   bufio.NewReader(r),
		closer:   r,
		metadata: &Metadata{},
	}
}

func (d *OpusDecoder) OpusFrame() ([]byte, error) {
	if d.closed {
		return nil, ErrDecoderClosed
	}

	if !d.firstFrameProcessed {
		if err := d.readMetadata(); err != nil {
			return nil, err
		}
	}
	return d.decodeFrame()
}

func (d *OpusDecoder) Close() error {
	if d.closed {
		return nil
	}
	d.closed = true
	if d.closer != nil {
		return d.closer.Close()
	}
	return nil
}

func (d *OpusDecoder) readMetadata() error {
	if d.firstFrameProcessed {
		return ErrNotFirstFrame
	}
	d.firstFrameProcessed = true

	header := make([]byte, 4)
	if _, err := io.ReadFull(d.reader, header); err != nil {
		return err
	}

	var metaLen int32
	if err := binary.Read(d.reader, binary.LittleEndian, &metaLen); err != nil {
		return err
	}

	if metaLen <= 0 {
		return ErrInvalidMetadata
	}

	jsonBuf := make([]byte, metaLen)
	if _, err := io.ReadFull(d.reader, jsonBuf); err != nil {
		return err
	}

	d.metadata = new(Metadata)
	return json.Unmarshal(jsonBuf, d.metadata)
}

func (d *OpusDecoder) decodeFrame() ([]byte, error) {
	var size int16
	if err := binary.Read(d.reader, binary.LittleEndian, &size); err != nil {
		return nil, err
	}

	if size < 0 {
		return nil, ErrNegativeFrameSize
	}

	frame := make([]byte, size)
	if _, err := io.ReadFull(d.reader, frame); err != nil {
		return nil, err
	}

	return frame, nil
}

func (d *OpusDecoder) frameDuration() time.Duration {
	if d.metadata == nil {
		return 20 * time.Millisecond
	}
	return time.Duration(((d.metadata.Opus.FrameSize/d.metadata.Opus.Channels)/960)*20) * time.Millisecond
}
