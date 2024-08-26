package encoder

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestEncode(t *testing.T) {
	ctx := context.Background()
	session, err := EncodeFile("deadpool-bye-bye.ogg", StdEncodeOptions, ctx)
	if err != nil {
		t.Fatal("Fallo al crear la session de encoding:", err)
	}

	outFile, err := os.Create("../decoder/deadpool-bye-bye.dca")
	if err != nil {
		t.Fatal("Error al crear el archivo:", err)
	}
	defer outFile.Close()

	numFrames := 0
	for {
		frame, err := session.ReadFrame()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatal("Fallo al leer los frames:", err)
		}
		_, err = outFile.Write(frame)
		if err != nil {
			t.Fatal("Fallo al escribir los frames en el archivo:", err)
		}
		numFrames++
	}

	expectedFrames := 11938
	if numFrames != expectedFrames {
		t.Errorf("Numero de frames incorrectos (obtenido %d, esperado %d)", numFrames, expectedFrames)
	}

	fmt.Println(session.FFMPEGMessages())
}
