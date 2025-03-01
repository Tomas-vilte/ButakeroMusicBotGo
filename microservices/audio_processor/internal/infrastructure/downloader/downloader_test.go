package downloader

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"testing"
	"time"
)

const (
	testTimeout = 30 * time.Second
	validURL    = "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	minFileSize = 1024 * 10
)

func setupTestDownloader(t *testing.T) (*YTDLPDownloader, logger.Logger) {
	testLogger, err := logger.NewZapLogger()
	require.NoError(t, err, "Error creando el logger")

	downloaderAudio, err := NewYTDLPDownloader(testLogger, YTDLPOptions{
		UseOAuth2: false,
	})
	require.NoError(t, err, "Error creando el downloader")

	return downloaderAudio, testLogger
}

func TestDownloadAudio(t *testing.T) {
	t.Skip("Omitir tests")
	t.Run("Create a new downloader", func(t *testing.T) {
		t.Run("Create successful downloader", func(t *testing.T) {
			testLogger, err := logger.NewZapLogger()
			require.NoError(t, err)

			_, err = NewYTDLPDownloader(testLogger, YTDLPOptions{
				UseOAuth2: false,
			})
			assert.NoError(t, err)
		})

		t.Run("Create with custom timeout", func(t *testing.T) {
			testLogger, err := logger.NewZapLogger()
			require.NoError(t, err)

			_, err = NewYTDLPDownloader(testLogger, YTDLPOptions{
				UseOAuth2: false,
			})
			assert.NoError(t, err)
		})
	})

	t.Run("Download audio", func(t *testing.T) {
		t.Run("Download successful", func(t *testing.T) {
			downloaderAudio, _ := setupTestDownloader(t)
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			reader, err := downloaderAudio.DownloadAudio(ctx, validURL)
			require.NoError(t, err)

			validateDownloadedContent(t, reader)
		})

		t.Run("Empty URL", func(t *testing.T) {
			downloaderAudio, _ := setupTestDownloader(t)
			ctx := context.Background()

			_, err := downloaderAudio.DownloadAudio(ctx, "")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "URL no puede ser vacía")
		})

		t.Run("Timeout exceeded", func(t *testing.T) {
			downloaderAudio, _ := setupTestDownloader(t)
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer cancel()

			_, err := downloaderAudio.DownloadAudio(ctx, validURL)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "context deadline exceeded")
		})

		t.Run("Download corrupted audio", func(t *testing.T) {
			downloaderAudio, _ := setupTestDownloader(t)
			ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
			defer cancel()

			reader, err := downloaderAudio.DownloadAudio(ctx, "https://www.youtube.com/watch?v=corrupt123")
			if err != nil {
				assert.Contains(t, err.Error(), "el archivo es muy chiquito")
				return
			}

			buf := make([]byte, 1024)
			n, _ := reader.Read(buf)
			assert.Less(t, n, minFileSize)
		})
	})
	t.Run("Handle context", func(t *testing.T) {
		t.Run("Context canceled", func(t *testing.T) {
			downloaderAudio, _ := setupTestDownloader(t)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			go func() {
				time.Sleep(100 * time.Microsecond)
				cancel()
			}()

			_, err := downloaderAudio.DownloadAudio(ctx, validURL)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "context canceled")
		})

		t.Run("Download concurrent", func(t *testing.T) {
			downloaderAudio, _ := setupTestDownloader(t)
			urls := []string{
				"https://www.youtube.com/watch?v=video1",
				"https://www.youtube.com/watch?v=video2",
				"https://www.youtube.com/watch?v=video3",
			}

			errChan := make(chan error, len(urls))

			for _, url := range urls {
				go func(url string) {
					ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
					defer cancel()

					reader, err := downloaderAudio.DownloadAudio(ctx, url)
					if err != nil {
						errChan <- err
						return
					}

					buf := make([]byte, 1024)
					_, err = reader.Read(buf)
					errChan <- err
				}(url)
			}

			for range urls {
				<-errChan
			}
		})
	})
}

func validateDownloadedContent(t *testing.T, reader io.Reader) {
	tmpFile, err := os.CreateTemp("", "downloaded-audio-*.m4a")
	require.NoError(t, err, "Error creando archivo temporal")

	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	size, err := io.Copy(tmpFile, reader)
	require.NoError(t, err, "Error copiando contenido descargado")

	assert.Greater(t, size, int64(minFileSize),
		fmt.Sprintf("El archivo descargado es muy pequeño: %d bytes", size))

	fileInfo, err := tmpFile.Stat()
	require.NoError(t, err, "Error obteniendo información del archivo")
	assert.NotEqual(t, int64(186), fileInfo.Size(),
		"El archivo tiene el tamaño típico de una descarga fallida (186 bytes)")
}
