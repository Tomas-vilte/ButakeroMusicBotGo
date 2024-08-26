package decoder

import (
	"io"
	"os"
	"testing"
)

func TestDecode(t *testing.T) {
	file, err := os.Open("deadpool-bye-bye.dca")
	if err != nil {
		t.Error(err)
	}

	decoder := NewDecoder(file)

	err = decoder.ReadMetadata()
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
