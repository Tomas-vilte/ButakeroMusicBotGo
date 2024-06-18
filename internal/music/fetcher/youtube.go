package fetcher

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/cache"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/internal/metrics"
	"go.uber.org/zap"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// SongLooker define la interfaz para buscar canciones.
type SongLooker interface {
	LookupSongs(ctx context.Context, input string) ([]*voice.Song, error)
	SearchYouTubeVideoID(ctx context.Context, searchTerm string) (string, error)
}

// YoutubeFetcher es un tipo que interactúa con YouTube para obtener metadatos y datos de audio.
type YoutubeFetcher struct {
	Logger        logging.Logger
	Cache         cache.Manager
	CacheMetrics  metrics.CacheMetrics
	youtubeAPIKey string
	audioCache    cache.AudioCaching
}

// NewYoutubeFetcher crea una nueva instancia de YoutubeFetcher con un logger predeterminado.
func NewYoutubeFetcher(logger logging.Logger, cache cache.Manager, cacheMetrics metrics.CacheMetrics, youtubeAPIKey string, audioCache cache.AudioCaching) *YoutubeFetcher {
	return &YoutubeFetcher{
		Logger:        logger,
		Cache:         cache,
		CacheMetrics:  cacheMetrics,
		youtubeAPIKey: youtubeAPIKey,
		audioCache:    audioCache,
	}
}

// LookupSongs busca canciones en YouTube según el término de búsqueda proporcionado en input.
// Retorna una lista de objetos bot.Song que contienen metadatos de las canciones encontradas.
func (s *YoutubeFetcher) LookupSongs(ctx context.Context, input string) ([]*voice.Song, error) {
	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", input)

	cachedResult := s.Cache.Get(videoURL)
	if cachedResult != nil {
		s.CacheMetrics.IncHits()
		return cachedResult, nil
	}

	video, err := s.getVideoDetails(ctx, input)
	if err != nil {
		s.Logger.Error("Error al obtener detalles del video", zap.Error(err))
		return nil, fmt.Errorf("error al obtener detalles del video")
	}

	duration, err := parseCustomDuration(video.ContentDetails.Duration)
	if err != nil {
		fmt.Println("Error al analizar la duración:", err)
	}
	thumbnailURL := video.Snippet.Thumbnails.Default.Url

	song := &voice.Song{
		Type:         "youtube",
		Title:        video.Snippet.Title,
		URL:          videoURL,
		Playable:     video.Snippet.LiveBroadcastContent != "live",
		ThumbnailURL: &thumbnailURL,
		Duration:     duration,
	}
	songs := []*voice.Song{song}

	s.Cache.Set(videoURL, songs)
	s.CacheMetrics.IncMisses()
	s.CacheMetrics.SetCacheSize(float64(s.Cache.Size()))

	return songs, nil
}

func parseCustomDuration(durationStr string) (time.Duration, error) {
	parts := strings.Split(durationStr, "T")
	if len(parts) != 2 {
		return 0, fmt.Errorf("formato de duración no válido: %s", durationStr)
	}

	durationPart := parts[1]
	var duration time.Duration

	for durationPart != "" {
		var value int64
		var unit string

		// Obtener el número y la unidad
		i := 0
		for ; i < len(durationPart); i++ {
			if durationPart[i] < '0' || durationPart[i] > '9' {
				break
			}
		}

		value, err := strconv.ParseInt(durationPart[:i], 10, 64)
		if err != nil {
			return 0, fmt.Errorf("formato de duración no válido: %s", durationStr)
		}

		unit = durationPart[i : i+1]

		// Actualizar la duración
		switch unit {
		case "H":
			duration += time.Duration(value) * time.Hour
		case "M":
			duration += time.Duration(value) * time.Minute
		case "S":
			duration += time.Duration(value) * time.Second
		default:
			return 0, fmt.Errorf("unidad de duración desconocida: %s", unit)
		}

		// Mover al siguiente componente de la duración
		durationPart = durationPart[i+1:]
	}

	return duration, nil
}

func (s *YoutubeFetcher) SearchYouTubeVideoID(ctx context.Context, searchTerm string) (string, error) {
	service, err := youtube.NewService(ctx, option.WithAPIKey(s.youtubeAPIKey))
	if err != nil {
		return "", fmt.Errorf("error al crear el cliente de la API de YouTube")
	}

	call := service.Search.List([]string{"id"}).Q(searchTerm).MaxResults(1).Type("video")
	response, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("error al buscar el video en YouTube: %w", err)
	}

	if len(response.Items) == 0 {
		return "", fmt.Errorf("no se encontró ningún video para el término de búsqueda: %s", searchTerm)
	}
	return response.Items[0].Id.VideoId, nil
}

func (s *YoutubeFetcher) getVideoDetails(ctx context.Context, videoID string) (*youtube.Video, error) {
	service, err := youtube.NewService(ctx, option.WithAPIKey(s.youtubeAPIKey))
	if err != nil {
		return nil, fmt.Errorf("error al crear el cliente de la API de YouTube")
	}

	call := service.Videos.List([]string{"snippet", "contentDetails", "liveStreamingDetails"}).Id(videoID)
	response, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("error al obtener los detalles del video: %w", err)
	}

	if len(response.Items) == 0 {
		return nil, fmt.Errorf("no se encontró el video con ID: %s", videoID)
	}

	return response.Items[0], nil
}

// GetDCAData obtiene los datos de audio de una canción en formato DCA.
// Utiliza yt-dlp y ffmpeg para descargar el audio de YouTube y convertirlo al formato DCA esperado por Discord.
// Retorna un io.Reader que permite leer los datos de audio y un posible error.
func (s *YoutubeFetcher) GetDCAData(ctx context.Context, song *voice.Song) (io.Reader, error) {
	startTime := time.Now()

	// Verificar si los datos de audio están en caché
	if cachedData, ok := s.audioCache.Get(song.URL); ok {
		s.CacheMetrics.IncRequests()
		return bytes.NewReader(cachedData), nil
	}

	// Descargar los datos de audio
	data, err := s.downloadAndCacheAudio(ctx, song)
	if err != nil {
		s.CacheMetrics.IncRequests()
		return nil, err
	}
	s.CacheMetrics.IncRequests()
	s.CacheMetrics.IncLatencyGet(time.Since(startTime))
	return bytes.NewReader(data), nil
}

func (s *YoutubeFetcher) downloadAndCacheAudio(ctx context.Context, song *voice.Song) ([]byte, error) {
	startTime := time.Now()
	ytArgs := []string{"-f", "bestaudio[ext=m4a]", "--audio-quality", "0", "-o", "-", "--force-overwrites", "--http-chunk-size", "100K", "'" + song.URL + "'"}
	ffmpegArgs := []string{"-i", "pipe:0", "-b:a", "192k", "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1"}

	// Ejecuta una cadena de comandos para descargar el audio de YouTube y convertirlo a formato DCA.
	downloadCmd := exec.CommandContext(ctx,
		"sh", "-c", fmt.Sprintf("yt-dlp %s | ffmpeg %s | dca",
			strings.Join(ytArgs, " "),
			strings.Join(ffmpegArgs, " ")))

	output, err := downloadCmd.Output()
	if err != nil {
		s.CacheMetrics.IncRequests()
		return nil, fmt.Errorf("error al ejecutar el comando: %w", err)
	}

	// Guardar los datos de audio en caché
	s.audioCache.Set(song.URL, output)
	s.CacheMetrics.IncRequests()
	s.CacheMetrics.IncLatencySet(time.Since(startTime))

	return output, nil
}
