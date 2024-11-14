package downloader

import (
	"bufio"
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"io"
	"os/exec"
	"strings"
	"sync"
)

type (
	// Downloader es una interfaz que define el contrato para descargar audio.
	Downloader interface {
		DownloadAudio(ctx context.Context, url string) (io.Reader, error)
	}

	// YTDLPDownloader es una implementación de Downloader que usa yt-dlp para descargar audio.
	YTDLPDownloader struct {
		log       logger.Logger
		useOAuth2 bool
	}

	// YTDLPOptions contiene las opciones de configuración para YTDLPDownloader.
	YTDLPOptions struct {
		UseOAuth2 bool
	}
)

// NewYTDLPDownloader crea y devuelve una nueva instancia de YTDLPDownloader.
func NewYTDLPDownloader(log logger.Logger, options YTDLPOptions) *YTDLPDownloader {
	return &YTDLPDownloader{
		log:       log,
		useOAuth2: options.UseOAuth2,
	}
}

// DownloadAudio implementa la interfaz Downloader para YTDLPDownloader.
// Descarga el audio de la URL proporcionada usando yt-dlp y devuelve un io.Reader para acceder al contenido.
func (d *YTDLPDownloader) DownloadAudio(ctx context.Context, url string) (io.Reader, error) {
	// Creamos un pipe para pasar el audio descargado
	pr, pw := io.Pipe()

	// Configuramos los argumentos para yt-dlp
	ytArgs := []string{
		"-f", "bestaudio[ext=m4a]",
		"--audio-quality", "0",
		"-o", "-",
		"--force-overwrites",
		"--http-chunk-size", "100K",
	}

	if d.useOAuth2 {
		ytArgs = append(ytArgs, "--username", "oauth", "--password", "''")
	}

	ytArgs = append(ytArgs, url)

	// Registramos el comando que se va a ejecutar
	d.log.Info("Ejecutando comando yt-dlp", zap.String("comando", fmt.Sprintf("yt-dlp %s", strings.Join(ytArgs, " "))))

	// Preparamos el comando con el contexto proporcionado
	cmd := exec.CommandContext(ctx, "yt-dlp", ytArgs...)

	// Configuramos los pipes para stdout y stderr
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error al crear el pipe de stdout: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("error al crear el pipe de stderr: %w", err)
	}

	// Conectamos la salida del comando al pipe de escritura
	cmd.Stdout = pw

	// Usamos un WaitGroup para sincronizar las goroutines de procesamiento de salida
	var wg sync.WaitGroup
	wg.Add(2)

	// Iniciamos goroutines para procesar stdout y stderr
	go d.processOutput(&wg, stdoutPipe, "stdout")
	go d.processOutput(&wg, stderrPipe, "stderr")

	// Iniciamos el comando
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error al iniciar el comando: %w", err)
	}

	// Goroutine para esperar que el comando termine y cerrar el pipe de escritura
	go func() {
		defer pw.Close()
		wg.Wait()
		if err := cmd.Wait(); err != nil {
			d.log.Error("error al esperar el comando", zap.Error(err))
		}
	}()

	return pr, nil
}

// processOutput maneja la salida de stdout o stderr del comando yt-dlp.
// Registra la salida usando el logger apropiado según el tipo y contenido del mensaje.
func (d *YTDLPDownloader) processOutput(wg *sync.WaitGroup, pipe io.ReadCloser, pipeType string) {
	defer wg.Done()

	reader := bufio.NewReader(pipe)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			d.log.Error(fmt.Sprintf("error leyendo %s", pipeType), zap.Error(err))
			break
		}

		line = strings.TrimSpace(line)

		if pipeType == "stdout" {
			if strings.Contains(line, "Downloading") || strings.Contains(line, "Progress:") {
				d.log.Info("yt-dlp progreso", zap.String("output", line))
			} else {
				d.log.Debug("yt-dlp stdout", zap.String("output", line))
			}
		} else {
			if strings.Contains(line, "WARNING") || strings.Contains(line, "ERROR") {
				d.log.Error("yt-dlp stderr", zap.String("error", line))
			} else {
				d.log.Info("yt-dlp stderr", zap.String("output", line))
			}
		}
	}
}
