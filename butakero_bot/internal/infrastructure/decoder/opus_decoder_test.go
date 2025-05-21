//go:build !integration

package decoder

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
)

type mockReadCloser struct {
	*bytes.Reader
	closed bool
}

func newMockReadCloser(data []byte) *mockReadCloser {
	return &mockReadCloser{
		Reader: bytes.NewReader(data),
		closed: false,
	}
}

func (m *mockReadCloser) Close() error {
	m.closed = true
	return nil
}

func TestDecode(t *testing.T) {
	file, err := os.Open("deadpool-bye-bye.dca")
	if err != nil {
		t.Error(err)
	}

	decoder := NewOpusDecoder(file)

	err = decoder.readMetadata()
	if err != nil {
		t.Error(err)
	}

	frameCounter := 0
	for {
		_, err := decoder.OpusFrame()
		if err != nil {
			if err != io.EOF {
				t.Error(err)
			}
			break
		}
		frameCounter++
	}

	if frameCounter != 11937 {
		t.Error("Numero de frames incorrectos")
	}
}

func TestOpusDecoder_Close(t *testing.T) {
	// Arrange
	mock := newMockReadCloser([]byte("test data"))
	decoder := NewOpusDecoder(mock)

	// Act
	err := decoder.Close()

	// Assert
	if err != nil {
		t.Errorf("Expected no error on close, got %v", err)
	}
	if !mock.closed {
		t.Error("Expected underlying reader to be closed")
	}

	frame, err := decoder.OpusFrame()
	if !errors.Is(err, ErrDecoderClosed) {
		t.Errorf("Expected ErrDecoderClosed after closing, got %v", err)
	}
	if frame != nil {
		t.Error("Expected nil frame after closing")
	}
}

func TestOpusDecoder_DoubleClose(t *testing.T) {
	// Arrange
	mock := newMockReadCloser([]byte("test data"))
	decoder := NewOpusDecoder(mock)

	// Act
	err1 := decoder.Close()
	err2 := decoder.Close()

	// Assert
	if err1 != nil || err2 != nil {
		t.Error("Expected no errors on multiple closes")
	}
}

func TestOpusDecoder_OpusFrame_EOF(t *testing.T) {
	// Arrange
	mock := newMockReadCloser([]byte{})
	decoder := NewOpusDecoder(mock)

	// Act
	frame, err := decoder.OpusFrame()

	// Assert
	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}
	if frame != nil {
		t.Error("Expected nil frame")
	}
}
