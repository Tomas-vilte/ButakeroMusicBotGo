package decoder

import (
	"bufio"
	"io"
	"os"
	"testing"
)

func BenchmarkOpusDecoder_WithBufio(b *testing.B) {
	file, err := os.Open("deadpool-bye-bye.dca")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			b.Fatal("Error cerrando el archivo")
		}
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = file.Seek(0, 0)
		bufferedFile := bufio.NewReader(file)

		decoder := &OpusDecoder{
			reader: bufferedFile,
			closer: file,
		}

		for {
			frame, err := decoder.OpusFrame()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
			_ = frame
		}
	}
}

func BenchmarkOpusDecoder_WithoutBufio(b *testing.B) {
	file, err := os.Open("deadpool-bye-bye.dca")
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			b.Fatal("Error cerrando el archivo")
		}
	}()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = file.Seek(0, 0)

		decoder := NewOpusDecoder(file)

		for {
			frame, err := decoder.OpusFrame()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.Fatal(err)
			}
			_ = frame
		}
	}
}
