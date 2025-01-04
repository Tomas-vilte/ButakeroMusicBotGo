package downloader

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
)

const (
	defaultTimeout   = 2 * time.Minute
	defaultChunkSize = "10M"
	audioFormat      = "bestaudio[ext=m4a]"
	audioQuality     = "0"
	minValidFileSize = 1024 * 10
)

type (
	// Downloader es una interfaz que define el contrato para descargar audio.
	Downloader interface {
		DownloadAudio(ctx context.Context, url string) (io.ReadCloser, error)
	}

	// YTDLPDownloader es una implementación de Downloader que usa yt-dlp para descargar audio.
	YTDLPDownloader struct {
		log       logger.Logger
		useOAuth2 bool
		cookies   string
		timeout   time.Duration
	}

	// YTDLPOptions contiene las opciones de configuración para YTDLPDownloader.
	YTDLPOptions struct {
		UseOAuth2 bool
		Cookies   string
		Timeout   time.Duration
	}

	downloadError struct {
		msg string
		err error
	}

	ValidationError struct {
		msg  string
		size int64
	}

	closeableReader struct {
		io.Reader
		closer func() error
	}
)

// NewYTDLPDownloader crea y devuelve una nueva instancia de YTDLPDownloader.
func NewYTDLPDownloader(log logger.Logger, options YTDLPOptions) (*YTDLPDownloader, error) {
	if log == nil {
		return nil, fmt.Errorf("el logger no puede ser nil")
	}

	if options.Timeout == 0 {
		options.Timeout = defaultTimeout
	}

	return &YTDLPDownloader{
		log:       log,
		useOAuth2: options.UseOAuth2,
		cookies:   options.Cookies,
		timeout:   options.Timeout,
	}, nil
}

// DownloadAudio implementa la interfaz Downloader para YTDLPDownloader.
// Descarga el audio de la URL proporcionada usando yt-dlp y devuelve un io.Reader para acceder al contenido.
func (d *YTDLPDownloader) DownloadAudio(ctx context.Context, url string) (io.ReadCloser, error) {
	if url == "" {
		return nil, fmt.Errorf("la URL no puede ser vacía")
	}

	ctx, cancel := context.WithTimeout(ctx, d.timeout)

	var buf bytes.Buffer
	pr, pw := io.Pipe()

	ytArgs := d.buildYTDLPArgs(url)
	d.log.Info("Arrancando yt-dlp", zap.Strings("args", ytArgs))

	cmd := exec.CommandContext(ctx, "yt-dlp", ytArgs...)

	stdoutPipe, stderrPipe, err := d.setupPipes(cmd)
	if err != nil {
		cancel()
		return nil, err
	}

	mw := io.MultiWriter(&buf, pw)
	cmd.Stdout = mw

	var downloadErr error
	var errBuf bytes.Buffer

	var wg sync.WaitGroup
	wg.Add(2)

	go d.processOutput(&wg, stdoutPipe, "stdout", cancel)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			line := scanner.Text()
			errBuf.WriteString(line)
			d.logOutput("stderr", line)
		}
	}()

	if err := cmd.Start(); err != nil {
		cancel()
		return nil, fmt.Errorf("error al iniciar yt-dlp: %w", err)
	}

	go func() {
		defer func() {
			if downloadErr != nil {
				pw.CloseWithError(downloadErr)
			} else {
				pw.Close()
			}
			cancel()
		}()

		wg.Wait()
		cmdErr := cmd.Wait()

		size := int64(buf.Len())
		if size < minValidFileSize {
			downloadErr = &ValidationError{
				msg:  "El archivo es muy chiquito, capaz que falló la descarga",
				size: size,
			}
			d.log.Error("La descarga falló mal",
				zap.Int64("Tamaño", size),
				zap.String("Error", errBuf.String()))
			return
		}

		if strings.Contains(errBuf.String(), "ERROR") {
			downloadErr = fmt.Errorf("yt-dlp se mando una cagada: %s", errBuf.String())
			return
		}

		if cmdErr != nil {
			downloadErr = fmt.Errorf("el comando yt-dlp falló: %w", cmdErr)
			return
		}
	}()

	return &closeableReader{
		Reader: pr,
		closer: func() error {
			cancel()
			return pr.Close()
		},
	}, nil
}

func (d *YTDLPDownloader) buildYTDLPArgs(url string) []string {
	args := []string{
		"-f", audioFormat,
		"--audio-quality", audioQuality,
		"-o", "-",
		"--force-overwrites",
		"--http-chunk-size", defaultChunkSize,
	}

	if d.cookies != "" {
		args = append(args, "--cookies", d.cookies)
	}

	args = append(args, url)
	return args
}

func (d *YTDLPDownloader) setupPipes(cmd *exec.Cmd) (io.ReadCloser, io.ReadCloser, error) {
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("error al crear el pipe de stdout: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("error al crear el pipe de stderr: %w", err)
	}

	return stdoutPipe, stderrPipe, nil
}

// processOutput maneja la salida de stdout o stderr del comando yt-dlp.
// Registra la salida usando el logger apropiado según el tipo y contenido del mensaje.
func (d *YTDLPDownloader) processOutput(wg *sync.WaitGroup, pipe io.ReadCloser, pipeType string, cancel context.CancelFunc) {
	defer wg.Done()
	reader := bufio.NewReader(pipe)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				d.log.Error("Hubo un error letyendo la salida del comando", zap.Error(err))
				cancel()
			}
			break
		}
		line = strings.TrimSpace(line)
		d.logOutput(pipeType, line)
	}
}

func (d *YTDLPDownloader) logOutput(pipeType, line string) {
	switch {
	case pipeType == "stdout" && (strings.Contains(line, "Downloading") || strings.Contains(line, "Progress:")):
		d.log.Info("Bajando cancion", zap.String("progreso", line))
	case pipeType == "stdout":
		d.log.Debug("Salida de yt-dlp", zap.String("info", line))
	case strings.Contains(line, "WARINIG"):
		d.log.Info("Advertencia de yt-dlp", zap.String("warning", line))
	case strings.Contains(line, "ERROR"):
		d.log.Error("Error de yt-dlp", zap.String("error", line))
	default:
		d.log.Info("Info de yt-dlp", zap.String("mensaje", line))
	}
}

func (e *downloadError) Error() string {
	return fmt.Sprintf("%s: %v", e.msg, e.err)
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s (tamaño: %d bytes)", e.msg, e.size)
}

func (c *closeableReader) Close() error {
	return c.closer()

}
