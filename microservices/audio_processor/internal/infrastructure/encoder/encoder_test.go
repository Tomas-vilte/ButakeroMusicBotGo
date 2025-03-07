//go:build !integration

package encoder

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestEncoder(t *testing.T) {
	// Preparar el contexto y el logger
	ctx := context.Background()
	logging, _ := logger.NewZapLogger()

	// Abrir el archivo de entrada .ogg
	inputFile, err := os.Open("./Twenty One Pilots - The Line (from Arcane Season 2) [Official Music Video].ogg")
	if err != nil {
		t.Fatalf("Error al abrir el archivo de entrada: %v", err)
	}
	defer inputFile.Close()

	sessionEncoder := NewFFmpegEncoder(logging)

	// Crear la sesión de codificación
	session, err := sessionEncoder.Encode(ctx, inputFile, StdEncodeOptions)
	if err != nil {
		t.Fatalf("Error al crear la sesión de codificación: %v", err)
	}
	defer session.Cleanup()

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error al obtener el directorio actual: %v", err)
	}

	// Crear el archivo de salida
	outPath := filepath.Join(currentDir, "test-song.dca")
	outFile, err := os.Create(outPath)
	if err != nil {
		t.Fatalf("Error al crear el archivo de salida: %v", err)
	}
	defer outFile.Close()

	// Leer frames y escribir en el archivo de salida
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

	// Verificar que se hayan procesado frames
	if numFrames == 0 {
		t.Error("No se procesaron frames")
	}

	// Verificar el tamaño del archivo de salida
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
