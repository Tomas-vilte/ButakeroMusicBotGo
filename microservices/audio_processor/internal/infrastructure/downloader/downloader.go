package downloader

import (
	"bufio"
	"context"
	"errors" // Importamos el paquete errors
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
		cookies   string
		errorChan chan error
	}

	// YTDLPOptions contiene las opciones de configuración para YTDLPDownloader.
	YTDLPOptions struct {
		UseOAuth2 bool
		Cookies   string
	}
)

// NewYTDLPDownloader crea y devuelve una nueva instancia de YTDLPDownloader.
func NewYTDLPDownloader(log logger.Logger, options YTDLPOptions) (*YTDLPDownloader, error) {
	if log == nil {
		return nil, fmt.Errorf("el logger no puede estar vacio")
	}
	return &YTDLPDownloader{
		log:       log,
		useOAuth2: options.UseOAuth2,
		cookies:   options.Cookies,
		errorChan: make(chan error, 1),
	}, nil
}

// DownloadAudio implementa la interfaz Downloader para YTDLPDownloader.
// Descarga el audio de la URL proporcionada usando yt-dlp y devuelve un io.Reader para acceder al contenido.
func (d *YTDLPDownloader) DownloadAudio(ctx context.Context, url string) (io.Reader, error) {
	d.log.Info("Iniciando descarga de audio", zap.String("url", url))

	pr, pw := io.Pipe()

	ytArgs := []string{
		"-f", "bestaudio",
		"--audio-quality", "0",
		"-o", "-",
		"--force-overwrites",
		"--http-chunk-size", "20M",
		"--newline",
	}

	if d.cookies != "" {
		ytArgs = append(ytArgs, "--cookies", d.cookies)
	}

	ytArgs = append(ytArgs, url)

	d.log.Info("Ejecutando comando yt-dlp", zap.String("comando", fmt.Sprintf("yt-dlp %s", strings.Join(ytArgs, " "))))

	cmd := exec.CommandContext(ctx, "yt-dlp", ytArgs...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("error al crear el pipe de stdout: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("error al crear el pipe de stderr: %w", err)
	}

	cmd.Stdout = pw

	var wg sync.WaitGroup
	wg.Add(2)

	go d.processOutput(&wg, stdoutPipe, "stdout")
	go d.processOutput(&wg, stderrPipe, "stderr")

	var cmdError error

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("error al iniciar el comando: %w", err)
	}

	go func() {
		defer func() {
			d.log.Debug("Cerrando el pipe de escritura pw")
			if err := pw.Close(); err != nil {
				d.errorChan <- fmt.Errorf("error al close pw: %w", err)
			}
		}()

		d.log.Debug("Esperando a que terminen las goroutines de procesamiento")
		wg.Wait()
		d.log.Debug("Goroutines de procesamiento terminadas")

		d.log.Debug("Esperando a que termine el comando")
		cmdError = cmd.Wait()
		if cmdError != nil {
			d.log.Error("Error al esperar el comando", zap.Error(cmdError))
		} else {
			d.log.Debug("El comando termino correctamente")
		}

		select {
		case stderrErr := <-d.errorChan:
			d.log.Debug("Error recibido del errorChan", zap.Error(stderrErr))
			if stderrErr != nil {
				d.log.Error("Error detectado en stderr", zap.Error(stderrErr))
				if err := pr.CloseWithError(stderrErr); err != nil {
					d.errorChan <- fmt.Errorf("error al close pr: %w", err)
				}
				return
			}
		default:
			d.log.Debug("No se recibio ningun error del errorChan")
		}
		if cmdError != nil {
			d.log.Error("cmdError detectado", zap.Error(cmdError))
			if err := pr.CloseWithError(cmdError); err != nil {
				d.errorChan <- fmt.Errorf("error al close pr: %w", err)
			}
		}

	}()

	return readAllAndReturnReader(pr, d.log)
}

func readAllAndReturnReader(r io.Reader, log logger.Logger) (io.Reader, error) {
	log.Debug("Iniciando readAllAndReturnReader")
	data, err := io.ReadAll(r)
	if err != nil {
		if errors.Is(err, io.ErrClosedPipe) {
			log.Debug("Error io.ErrClosedPipe detectado en readAllAndReturnReader", zap.Error(err))
			return nil, err
		}
		log.Error("Error al leer reader en readAllAndReturnReader", zap.Error(err))
		return nil, fmt.Errorf("error al leer el reader: %w", err)
	}

	log.Debug("Creando nuevo io.Reader en readAllAndReturnReader")
	return strings.NewReader(string(data)), nil
}

// processOutput maneja la salida de stdout o stderr del comando yt-dlp.
// Registra la salida usando el logger apropiado según el tipo y contenido del mensaje.
func (d *YTDLPDownloader) processOutput(wg *sync.WaitGroup, pipe io.ReadCloser, pipeType string) {
	defer wg.Done()
	d.log.Debug("Iniciando processOutput", zap.String("pipeType", pipeType))

	scanner := bufio.NewScanner(pipe)
	var stderrLines []string

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		if pipeType == "stdout" {
			if strings.Contains(line, "Downloading") || strings.Contains(line, "Progress:") {
				d.log.Info("yt-dlp progreso", zap.String("output", line))
			} else {
				d.log.Debug("yt-dlp stdout", zap.String("output", line))
			}
		} else if pipeType == "stderr" {
			if strings.Contains(line, "WARNING") || strings.Contains(line, "ERROR") {
				d.log.Error("yt-dlp stderr", zap.String("error", line))
				stderrLines = append(stderrLines, line)
			} else {
				d.log.Info("yt-dlp stderr", zap.String("output", line))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		d.log.Error(fmt.Sprintf("error leyendo %s", pipeType), zap.Error(err))
	}

	if pipeType == "stderr" && len(stderrLines) > 0 {
		errorString := strings.Join(stderrLines, "\n")
		err := errors.New(errorString)
		d.log.Debug("Enviando error al errorChan", zap.Error(err))
		d.errorChan <- err
	}

	d.log.Debug("Finalizando processOutput", zap.String("pipeType", pipeType))
}
