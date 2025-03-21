package downloader

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	errorsApp "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/errors"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"io"
	"os/exec"
	"strings"
	"sync"
)

type (
	YTDLPDownloader struct {
		log       logger.Logger
		cookies   string
		errorChan chan error
	}

	YTDLPOptions struct {
		Cookies string
	}
)

func NewYTDLPDownloader(log logger.Logger, options YTDLPOptions) (*YTDLPDownloader, error) {
	if log == nil {
		return nil, errorsApp.ErrInvalidInput.WithMessage("el logger no puede estar vacio")
	}
	return &YTDLPDownloader{
		log:       log,
		cookies:   options.Cookies,
		errorChan: make(chan error, 1),
	}, nil
}

func (d *YTDLPDownloader) DownloadAudio(ctx context.Context, url string) (io.Reader, error) {
	log := d.log.With(
		zap.String("component", "YTDLPDownloader"),
		zap.String("method", "DownloadAudio"),
	)

	log.Info("Iniciando descarga de audio",
		zap.String("url", url),
		zap.String("cookies", d.cookies),
	)

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

	log.Debug("Ejecutando comando yt-dlp",
		zap.String("comando", fmt.Sprintf("yt-dlp %s", strings.Join(ytArgs, " "))),
	)

	cmd := exec.CommandContext(ctx, "yt-dlp", ytArgs...)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errorsApp.ErrYTDLPCommandFailed.WithMessage(fmt.Sprintf("error al crear el pipe de stdout: %v", err))
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, errorsApp.ErrYTDLPCommandFailed.WithMessage(fmt.Sprintf("error al crear el pipe de stderr: %v", err))
	}

	cmd.Stdout = pw

	var wg sync.WaitGroup
	wg.Add(2)

	go d.processOutput(&wg, stdoutPipe, "stdout")
	go d.processOutput(&wg, stderrPipe, "stderr")

	var cmdError error

	if err := cmd.Start(); err != nil {
		return nil, errorsApp.ErrYTDLPCommandFailed.WithMessage(fmt.Sprintf("error al iniciar el comando: %v", err))
	}

	go func() {
		defer func() {
			log.Debug("Cerrando el pipe de escritura",
				zap.String("pipeType", "pw"),
			)
			if err := pw.Close(); err != nil {
				log.Error("Error al cerrar el pipe de escritura",
					zap.Error(err),
				)
			}
		}()

		log.Debug("Esperando a que terminen las goroutines de procesamiento")
		wg.Wait()
		log.Debug("Goroutines de procesamiento terminadas")

		log.Debug("Esperando a que termine el comando")
		cmdError = cmd.Wait()
		if cmdError != nil {
			log.Error("Error al ejecutar el comando yt-dlp",
				zap.Error(cmdError),
				zap.String("comando", fmt.Sprintf("yt-dlp %s", strings.Join(ytArgs, " "))),
			)
		} else {
			log.Debug("El comando terminó correctamente")
		}

		select {
		case stderrErr := <-d.errorChan:
			log.Debug("Error recibido del errorChan",
				zap.Error(stderrErr),
			)
			if stderrErr != nil {
				log.Error("Error detectado en stderr",
					zap.Error(stderrErr),
				)
				if err := pr.CloseWithError(stderrErr); err != nil {
					log.Error("Error al cerrar el pipe de lectura",
						zap.Error(err),
					)
				}
				return
			}
		default:
			log.Debug("No se recibió ningún error del errorChan")
		}

		if cmdError != nil {
			log.Error("Error en la ejecución del comando",
				zap.Error(cmdError),
			)
			if err := pr.CloseWithError(cmdError); err != nil {
				log.Error("Error al cerrar el pipe de lectura",
					zap.Error(err),
				)
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
		return nil, errorsApp.ErrYTDLPInvalidOutput.WithMessage(fmt.Sprintf("error al leer el reader: %v", err))
	}

	log.Debug("Creando nuevo io.Reader en readAllAndReturnReader")
	return strings.NewReader(string(data)), nil
}

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
				d.log.Info("Progreso de yt-dlp",
					zap.String("pipeType", pipeType),
					zap.String("output", line),
				)
			} else {
				d.log.Debug("Salida de yt-dlp",
					zap.String("pipeType", pipeType),
					zap.String("output", line),
				)
			}
		} else if pipeType == "stderr" {
			if strings.Contains(line, "WARNING") || strings.Contains(line, "ERROR") {
				d.log.Error("Error en yt-dlp",
					zap.String("pipeType", pipeType),
					zap.String("error", line),
				)
				stderrLines = append(stderrLines, line)
			} else {
				d.log.Info("Salida de yt-dlp",
					zap.String("pipeType", pipeType),
					zap.String("output", line),
				)
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
