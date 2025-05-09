//go:build !integration

package encoder

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestEncoder(t *testing.T) {
	ctx := context.Background()
	logging, _ := logger.NewDevelopmentLogger()

	inputFile, err := os.Open("./Twenty One Pilots - The Line (from Arcane Season 2) [Official Music Video].ogg")
	if err != nil {
		t.Fatalf("Error al abrir el archivo de entrada: %v", err)
	}
	defer func() {
		if err := inputFile.Close(); err != nil {
			t.Fatalf("Error al cerrar el archivo de entrada: %v", err)
		}
	}()

	sessionEncoder := NewFFmpegEncoder(logging)

	session, err := sessionEncoder.Encode(ctx, inputFile, model.StdEncodeOptions)
	if err != nil {
		t.Fatalf("Error al crear la sesión de codificación: %v", err)
	}
	defer session.Cleanup()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error al obtener el directorio actual: %v", err)
	}

	outPath := filepath.Join(currentDir, "test-song.dca")
	outFile, err := os.Create(outPath)
	if err != nil {
		t.Fatalf("Error al crear el archivo de salida: %v", err)
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			t.Fatalf("Error al cerrar el archivo de salida: %v", err)
		}
	}()

	numFrames := 0
	for {
		frame, err := session.ReadFrame()
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("Error al leer frames: %v", err)
		}
		_, err = outFile.Write(frame)
		if err != nil {
			t.Fatalf("Error al escribir frames en el archivo: %v", err)
		}
		numFrames++
	}

	t.Logf("Número de frames procesados: %d", numFrames)

	if numFrames == 0 {
		t.Error("No se procesaron frames")
	}

	outInfo, err := outFile.Stat()
	if err != nil {
		t.Fatalf("Error al obtener información del archivo de salida: %v", err)
	}
	if outInfo.Size() == 0 {
		t.Error("El archivo de salida está vacío")
	}
	t.Logf("Archivo DCA generado en: %s", outPath)

	t.Logf("Mensajes de FFmpeg: %v", session.FFMPEGMessages())
}
