package downloader

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/music_download/internal/uploader"
	"go.uber.org/zap"
	"os/exec"
	"strings"
)

// Downloader es la interfaz para descargar canciones
type Downloader interface {
	DownloadSong(songURL string, key string) error
}

// YtDlpDownloader implementa la interfaz Downloader usando yt-dlp
type YtDlpDownloader struct {
	S3Uploader uploader.Uploader
	Logger     logging.Logger
	Executor   CommandExecutor
	YtPath     string
}

// CommandExecutor define una interfaz para ejecutar comandos del sistema.
type CommandExecutor interface {
	ExecuteCommand(ctx context.Context, name string, args ...string) ([]byte, error)
}

// DefaultCommandExecutor es una implementación concreta de CommandExecutor
type DefaultCommandExecutor struct{}

func NewCommandExecutor() *DefaultCommandExecutor {
	return &DefaultCommandExecutor{}
}

// ExecuteCommand ejecuta un comando del sistema y retorna la salida y el error
func (e *DefaultCommandExecutor) ExecuteCommand(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

func NewDownloader(s3uploader uploader.Uploader, logger logging.Logger, executor CommandExecutor, ytPath string) *YtDlpDownloader {
	return &YtDlpDownloader{
		S3Uploader: s3uploader,
		Logger:     logger,
		Executor:   executor,
		YtPath:     ytPath,
	}
}

// DownloadSong descarga una canción usando yt-dlp con los argumentos especificados
func (d *YtDlpDownloader) DownloadSong(songURL string, key string) error {
	ytArgs := []string{
		"-f", "bestaudio[ext=m4a]",
		"--audio-quality", "0",
		"-o", "-",
		"--force-overwrites",
		"--http-chunk-size", "100K",
		songURL,
	}

	output, err := d.Executor.ExecuteCommand(context.Background(), "sh", "-c", fmt.Sprintf("%s %s",
		d.YtPath, strings.Join(ytArgs, " ")))
	if err != nil {
		d.Logger.Error("Error al ejecutar yt-dlp", zap.Error(err))
		return fmt.Errorf("error al ejecutar yt-dlp: %v", err)
	}

	err = d.S3Uploader.UploadToS3(context.Background(), bytes.NewReader(output), key)
	if err != nil {
		d.Logger.Error("Error al subir datos a s3", zap.Error(err))
		return fmt.Errorf("error al subir datos a S3: %v", err)
	}
	return nil
}
