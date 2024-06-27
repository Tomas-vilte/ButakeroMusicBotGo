package fetcher

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/youtube/v3"
	"io"
	"os/exec"
	"testing"
	"time"
)

func TestYoutubeFetcher_LookupSongs(t *testing.T) {
	t.Run("Normal case", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockLogger)
		mockCache := new(MockCacheManager)
		mockYoutubeService := new(MockYouTubeService)
		mockAudioCache := new(MockAudioCaching)
		mockCommandExecutor := new(MockCommandExecutor)

		fetcher := NewYoutubeFetcher(mockLogger, mockCache, mockYoutubeService, mockAudioCache, mockCommandExecutor)

		ctx := context.Background()
		input := "dQw4w9WgXcQ"
		videoURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
		thumbnailURL := "https://i3.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg"

		expectedSong := &voice.Song{
			Type:         "youtube_provider",
			Title:        "Rick Astley - Never Gonna Give You Up (Official Music Video)",
			URL:          videoURL,
			Playable:     true,
			ThumbnailURL: &thumbnailURL,
			Duration:     time.Minute*3 + time.Second*33,
		}

		mockCache.On("Get", videoURL).Return(nil)
		mockYoutubeService.On("GetVideoDetails", ctx, input).Return(&youtube.Video{
			Snippet: &youtube.VideoSnippet{
				Title:                "Rick Astley - Never Gonna Give You Up (Official Music Video)",
				LiveBroadcastContent: "None",
				Thumbnails: &youtube.ThumbnailDetails{
					Default: &youtube.Thumbnail{
						Url: thumbnailURL,
					},
				},
			},
			ContentDetails: &youtube.VideoContentDetails{
				Duration: "PT3M33S",
			},
		}, nil)
		mockCache.On("Set", videoURL, []*voice.Song{expectedSong})

		// Act
		songs, err := fetcher.LookupSongs(ctx, input)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, songs, 1)
		assert.Equal(t, expectedSong.Title, songs[0].Title)
		assert.Equal(t, expectedSong.URL, songs[0].URL)
		assert.Equal(t, expectedSong.Playable, songs[0].Playable)
		assert.Equal(t, expectedSong.Duration, songs[0].Duration)
		assert.Equal(t, expectedSong.ThumbnailURL, songs[0].ThumbnailURL)

		mockCache.AssertExpectations(t)
		mockYoutubeService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Error fetching video details", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockLogger)
		mockCache := new(MockCacheManager)
		mockYoutubeService := new(MockYouTubeService)
		mockAudioCache := new(MockAudioCaching)
		mockCommandExecutor := new(MockCommandExecutor)

		fetcher := NewYoutubeFetcher(mockLogger, mockCache, mockYoutubeService, mockAudioCache, mockCommandExecutor)

		ctx := context.Background()
		input := "invalidVideoID"
		videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", input)

		expectedError := fmt.Errorf("error al obtener detalles del video")

		mockCache.On("Get", videoURL).Return(nil)
		mockYoutubeService.On("GetVideoDetails", ctx, input).Return(&youtube.Video{}, expectedError)
		mockLogger.On("Error", "Error al obtener detalles del video", mock.Anything)

		// Act
		songs, err := fetcher.LookupSongs(ctx, input)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, songs)
		assert.Equal(t, expectedError.Error(), err.Error())

		mockCache.AssertExpectations(t)
		mockYoutubeService.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("Cached result", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockLogger)
		mockCache := new(MockCacheManager)
		mockYoutubeService := new(MockYouTubeService)
		mockAudioCache := new(MockAudioCaching)
		mockCommandExecutor := new(MockCommandExecutor)

		fetcher := NewYoutubeFetcher(mockLogger, mockCache, mockYoutubeService, mockAudioCache, mockCommandExecutor)

		ctx := context.Background()
		input := "dQw4w9WgXcQ"
		videoURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
		thumbnailURL := "https://i3.ytimg.com/vi/dQw4w9WgXcQ/maxresdefault.jpg"
		expectedSong := &voice.Song{
			Type:         "youtube_provider",
			Title:        "Rick Astley - Never Gonna Give You Up (Official Music Video)",
			URL:          videoURL,
			Playable:     true,
			ThumbnailURL: &thumbnailURL,
			Duration:     time.Minute*3 + time.Second*33,
		}
		expectedCacheResult := []*voice.Song{expectedSong}

		mockCache.On("Get", videoURL).Return(expectedCacheResult)
		mockLogger.On("Info", "Video encontrado en cache: ", mock.AnythingOfType("[]zapcore.Field"))

		// Act
		songs, err := fetcher.LookupSongs(ctx, input)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, songs, 1)
		assert.Equal(t, expectedSong.Title, songs[0].Title)
		assert.Equal(t, expectedSong.URL, songs[0].URL)
		assert.Equal(t, expectedSong.Playable, songs[0].Playable)
		assert.Equal(t, expectedSong.Duration, songs[0].Duration)
		assert.Equal(t, expectedSong.ThumbnailURL, songs[0].ThumbnailURL)

		mockCache.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
		mockYoutubeService.AssertNotCalled(t, "GetVideoDetails", mock.Anything, mock.Anything)
	})
}

func TestYoutubeFetcher_GetDCAData(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockLogger)
		mockCache := new(MockCacheManager)
		mockYoutubeService := new(MockYouTubeService)
		mockAudioCache := new(MockAudioCaching)
		mockCommandExecutor := new(MockCommandExecutor)

		fetcher := NewYoutubeFetcher(mockLogger, mockCache, mockYoutubeService, mockAudioCache, mockCommandExecutor)

		ctx := context.Background()
		song := &voice.Song{
			URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		}

		// Crear un exec.Cmd mockeado para simular la ejecución de comandos
		cmd := exec.CommandContext(ctx, "echo", "fake audio data")
		mockCommandExecutor.On("ExecuteCommand", ctx, "sh", mock.Anything).Return(cmd)
		mockAudioCache.On("Get", song.URL).Return(nil, false)
		mockAudioCache.On("Set", song.URL, mock.Anything)

		// Act
		reader, err := fetcher.GetDCAData(ctx, song)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, reader)

		// Leer los datos del reader
		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, reader)
		require.NoError(t, err)

		// Verificar que los datos leídos son correctos (simulación de datos de audio)
		assert.Equal(t, "fake audio data\n", buf.String())

		mockCommandExecutor.AssertExpectations(t)
		mockAudioCache.AssertExpectations(t)
	})

	t.Run("Error executing command", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockLogger)
		mockCache := new(MockCacheManager)
		mockYoutubeService := new(MockYouTubeService)
		mockAudioCache := new(MockAudioCaching)
		mockCommandExecutor := new(MockCommandExecutor)

		fetcher := NewYoutubeFetcher(mockLogger, mockCache, mockYoutubeService, mockAudioCache, mockCommandExecutor)

		ctx := context.Background()
		song := &voice.Song{
			URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		}

		// Simular un comando que falla
		failingCmd := exec.Command("false")
		mockCommandExecutor.On("ExecuteCommand", ctx, "sh", mock.Anything).Return(failingCmd)

		mockAudioCache.On("Get", song.URL).Return(nil, false)
		mockLogger.On("Error", "Error al descargar y transmitir audio", mock.Anything)

		// Act
		reader, err := fetcher.GetDCAData(ctx, song)

		// Assert
		assert.NoError(t, err) // GetDCAData no devuelve error directamente
		assert.NotNil(t, reader)

		// Leer del reader para provocar el error
		_, readErr := io.ReadAll(reader)
		assert.Error(t, readErr)
		assert.Contains(t, readErr.Error(), "exit status 1")

		mockCommandExecutor.AssertExpectations(t)
		mockAudioCache.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("CachedDataAvailable", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockLogger)
		mockCache := new(MockCacheManager)
		mockYoutubeService := new(MockYouTubeService)
		mockAudioCache := new(MockAudioCaching)
		mockCommandExecutor := new(MockCommandExecutor)

		fetcher := NewYoutubeFetcher(mockLogger, mockCache, mockYoutubeService, mockAudioCache, mockCommandExecutor)

		ctx := context.Background()
		song := &voice.Song{
			URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		}
		cachedData := []byte("fake cached audio data")

		mockAudioCache.On("Get", song.URL).Return(cachedData, true)

		// Act
		reader, err := fetcher.GetDCAData(ctx, song)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, reader)

		buffer := make([]byte, len(cachedData))
		n, err := reader.Read(buffer)
		assert.NoError(t, err)
		assert.Equal(t, len(cachedData), n)
		assert.Equal(t, cachedData, buffer)

		mockAudioCache.AssertExpectations(t)
	})

	t.Run("CachedDataNotAvailable", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockLogger)
		mockCache := new(MockCacheManager)
		mockYoutubeService := new(MockYouTubeService)
		mockAudioCache := new(MockAudioCaching)
		mockCommandExecutor := new(MockCommandExecutor)

		fetcher := NewYoutubeFetcher(mockLogger, mockCache, mockYoutubeService, mockAudioCache, mockCommandExecutor)

		ctx := context.Background()
		song := &voice.Song{
			URL: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		}

		mockAudioCache.On("Get", song.URL).Return(nil, false)

		// Simular un comando que produce datos de audio
		fakeAudioData := []byte("fake audio data")
		cmd := exec.Command("echo", "-n", string(fakeAudioData))
		mockCommandExecutor.On("ExecuteCommand", ctx, "sh", mock.Anything).Return(cmd)

		mockAudioCache.On("Set", song.URL, mock.Anything).Run(func(args mock.Arguments) {
			// Verificar que los datos almacenados en caché son correctos
			cachedData := args.Get(1).([]byte)
			assert.Equal(t, fakeAudioData, cachedData)
		})

		// Act
		reader, err := fetcher.GetDCAData(ctx, song)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, reader)

		// Leer los datos del reader y verificar que son correctos
		data, readErr := io.ReadAll(reader)
		assert.NoError(t, readErr)
		assert.Equal(t, fakeAudioData, data)

		mockAudioCache.AssertExpectations(t)
		mockCommandExecutor.AssertExpectations(t)
	})
}

func TestYoutubeFetcher_SearchYouTubeVideoID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockLogger)
		mockCache := new(MockCacheManager)
		mockYoutubeService := new(MockYouTubeService)
		mockAudioCache := new(MockAudioCaching)
		mockCommandExecutor := new(MockCommandExecutor)

		fetcher := NewYoutubeFetcher(mockLogger, mockCache, mockYoutubeService, mockAudioCache, mockCommandExecutor)

		ctx := context.Background()
		searchTerm := "Rick Astley Never Gonna Give You Up"
		expectedVideoID := "dQw4w9WgXcQ"

		mockYoutubeService.On("SearchVideoID", ctx, searchTerm).Return(expectedVideoID, nil)

		// Act
		videoID, err := fetcher.SearchYouTubeVideoID(ctx, searchTerm)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, expectedVideoID, videoID)

		mockYoutubeService.AssertExpectations(t)
	})

	t.Run("Error", func(t *testing.T) {
		// Arrange
		mockLogger := new(MockLogger)
		mockCache := new(MockCacheManager)
		mockYoutubeService := new(MockYouTubeService)
		mockAudioCache := new(MockAudioCaching)
		mockCommandExecutor := new(MockCommandExecutor)

		fetcher := NewYoutubeFetcher(mockLogger, mockCache, mockYoutubeService, mockAudioCache, mockCommandExecutor)

		ctx := context.Background()
		searchTerm := "Rick Astley Never Gonna Give You Up"
		expectedError := fmt.Errorf("error buscando el video en YouTube")

		mockYoutubeService.On("SearchVideoID", ctx, searchTerm).Return("", expectedError)

		// Act
		videoID, err := fetcher.SearchYouTubeVideoID(ctx, searchTerm)

		// Assert
		assert.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("error al buscar el video en YouTube: %v", expectedError))
		assert.Empty(t, videoID)

		mockYoutubeService.AssertExpectations(t)
	})
}
