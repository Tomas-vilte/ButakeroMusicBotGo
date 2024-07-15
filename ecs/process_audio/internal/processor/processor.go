package processor

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/ecs/process_audio/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/ecs/process_audio/internal/uploader"
	"go.uber.org/zap"
	"io"
	"os/exec"
)

type (
	AudioProcessor interface {
		ProcessToDCA(ctx context.Context, key, inputFile string) error
	}

	FfmpegAudioProcessor struct {
		executor   CommandExecutor
		ffmpegPath string
		dcaPath    string
		s3Client   uploader.Uploader
		logger     logging.Logger
	}

	// CommandExecutor define una interfaz para ejecutar comandos del sistema.
	CommandExecutor interface {
		ExecuteCommand(ctx context.Context, name string, args ...string) *exec.Cmd
	}

	DefaultCommandExecutor struct{}
)

func NewAudioProcessor(logger logging.Logger, executor CommandExecutor, s3Client uploader.Uploader, dcaPath, ffmpeg string) *FfmpegAudioProcessor {
	return &FfmpegAudioProcessor{
		executor:   executor,
		ffmpegPath: ffmpeg,
		dcaPath:    dcaPath,
		s3Client:   s3Client,
		logger:     logger,
	}
}

// ExecuteCommand ejecuta un comando del sistema y retorna la salida y el error
func (e *DefaultCommandExecutor) ExecuteCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

func NewCommandExecutor() *DefaultCommandExecutor {
	return &DefaultCommandExecutor{}
}

func (p *FfmpegAudioProcessor) ProcessToDCA(ctx context.Context, key, inputFile string) error {
	// Descargar el archivo desde S3
	s3Object, err := p.s3Client.DownloadFromS3(ctx, inputFile)
	if err != nil {
		p.logger.Error("Error al descargar el archivo de S3", zap.Error(err), zap.String("inputFile", inputFile))
		return fmt.Errorf("error al descargar el archivo de S3: %w", err)
	}
	defer s3Object.Body.Close()

	// Leer el contenido del archivo en un buffer
	buffer := new(bytes.Buffer)
	_, err = io.Copy(buffer, s3Object.Body)
	if err != nil {
		p.logger.Error("Error al leer el archivo de S3 en el buffer", zap.Error(err))
		return fmt.Errorf("error al leer el archivo de S3 en el buffer: %w", err)
	}

	// Configurar el comando FFmpeg
	ffmpegCmd := p.executor.ExecuteCommand(ctx, p.ffmpegPath,
		"-i", "pipe:0",
		"-b:a", "192k",
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"pipe:1")
	ffmpegCmd.Stdin = buffer
	var ffmpegStderr bytes.Buffer
	ffmpegCmd.Stderr = &ffmpegStderr

	// Configurar el comando DCA
	dcaCmd := p.executor.ExecuteCommand(ctx, p.dcaPath)
	var dcaStderr bytes.Buffer
	dcaCmd.Stderr = &dcaStderr

	// Crear pipes para conectar FFmpeg y DCA
	ffmpegStdout, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error al crear pipe para FFmpeg stdout: %w", err)
	}
	dcaStdin, err := dcaCmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("error al crear pipe para DCA stdin: %w", err)
	}
	dcaStdout, err := dcaCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error al crear pipe para DCA stdout: %w", err)
	}

	// Iniciar los comandos
	p.logger.Info("Iniciando FFmpeg", zap.Strings("args", ffmpegCmd.Args))
	if err := ffmpegCmd.Start(); err != nil {
		return fmt.Errorf("error al iniciar FFmpeg: %w", err)
	}
	p.logger.Info("Iniciando DCA", zap.String("path", p.dcaPath))
	if err := dcaCmd.Start(); err != nil {
		return fmt.Errorf("error al iniciar DCA: %w", err)
	}

	// Copiar la salida de FFmpeg a la entrada de DCA
	go func() {
		_, err := io.Copy(dcaStdin, ffmpegStdout)
		if err != nil {
			p.logger.Error("Error al copiar datos de FFmpeg a DCA", zap.Error(err))
		}
		dcaStdin.Close()
	}()

	// Leer la salida de DCA
	output, err := io.ReadAll(dcaStdout)
	if err != nil {
		p.logger.Error("Error al leer la salida de DCA",
			zap.Error(err),
			zap.String("stderr", dcaStderr.String()))
		return fmt.Errorf("error al leer la salida de DCA: %w", err)
	}

	// Esperar a que ambos comandos terminen
	if err := ffmpegCmd.Wait(); err != nil {
		p.logger.Error("Error en FFmpeg",
			zap.Error(err),
			zap.String("stderr", ffmpegStderr.String()))
		return fmt.Errorf("error en FFmpeg: %w", err)
	}
	if err := dcaCmd.Wait(); err != nil {
		p.logger.Error("Error en DCA",
			zap.Error(err),
			zap.String("stderr", dcaStderr.String()))
		return fmt.Errorf("error en DCA: %w", err)
	}

	// Subir el archivo procesado a S3
	err = p.s3Client.UploadToS3(ctx, bytes.NewReader(output), key)
	if err != nil {
		p.logger.Error("Error al subir datos a S3",
			zap.Error(err),
			zap.String("key", key),
			zap.Int("dataSize", len(output)))
		return fmt.Errorf("error al subir datos a S3: %w", err)
	}

	p.logger.Info("Archivo procesado y subido exitosamente",
		zap.String("key", key),
		zap.Int("outputSize", len(output)))
	return nil
}
