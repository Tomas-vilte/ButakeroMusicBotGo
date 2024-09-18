package unit

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/downloader"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDownloadAudio(t *testing.T) {
	testLogger, err := logger.NewZapLogger()
	require.NoError(t, err)
	downloaderAudio := downloader.NewYTDLPDownloader(testLogger, downloader.YTDLPOptions{UseOAuth2: false})

	reader, err := downloaderAudio.DownloadAudio(context.TODO(), "https://www.youtube.com/watch?v=dQw4w9WgXcQ")

	assert.NoError(t, err, "La descarga de audio debería completarse sin errores")

	buf := make([]byte, 1024)
	n, err := reader.Read(buf)

	assert.NoError(t, err, "No debería haber error al leer del buffer")
	assert.Greater(t, n, 0, "El buffer debería contener datos")
}
