package fetcher

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/cache"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/internal/services/providers"
	"github.com/Tomas-vilte/GoMusicBot/internal/storage/s3_audio"
	"go.uber.org/zap"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type (
	// SongLooker define la interfaz para buscar canciones.
	SongLooker interface {
		LookupSongs(ctx context.Context, input string) ([]*voice.Song, error)
		SearchYouTubeVideoID(ctx context.Context, searchTerm string) (string, error)
	}

	// YoutubeFetcher es un tipo que interactúa con YouTube para obtener metadatos y datos de audio.
	YoutubeFetcher struct {
		Logger          logging.Logger
		Cache           cache.Manager
		audioCache      cache.AudioCaching
		YoutubeService  providers.YouTubeService
		CommandExecutor CommandExecutor
		S3Uploader      s3_audio.Uploader

		// Esto es para uso temporal! Debido a que youtube pide oauth, ademas con esto podemos evitar baneamiento de IP
		//username string
		//password string
	}

	// CommandExecutor define una interfaz para ejecutar comandos del sistema.
	CommandExecutor interface {
		ExecuteCommand(ctx context.Context, name string, args ...string) *exec.Cmd
	}

	DefaultCommandExecutor struct{}
)

func (e *DefaultCommandExecutor) ExecuteCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

func NewCommandExecutor() *DefaultCommandExecutor {
	return &DefaultCommandExecutor{}
}

// NewYoutubeFetcher crea una nueva instancia de YoutubeFetcher con un logger predeterminado.
func NewYoutubeFetcher(logger logging.Logger, cache cache.Manager, youtubeService providers.YouTubeService, audioCache cache.AudioCaching, commandExecutor CommandExecutor, s3Upload s3_audio.Uploader) *YoutubeFetcher {
	return &YoutubeFetcher{
		Logger:          logger,
		Cache:           cache,
		YoutubeService:  youtubeService,
		audioCache:      audioCache,
		CommandExecutor: commandExecutor,
		S3Uploader:      s3Upload,
	}
}

// LookupSongs busca canciones en YouTube según el término de búsqueda proporcionado en input.
// Retorna una lista de objetos voice.Song que contienen metadatos de las canciones encontradas.
func (s *YoutubeFetcher) LookupSongs(ctx context.Context, input string) ([]*voice.Song, error) {
	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", input)

	cachedResult := s.Cache.Get(videoURL)
	if cachedResult != nil {
		s.Logger.Info("Video encontrado en cache: ", zap.String("Video", videoURL))
		return cachedResult, nil
	}

	video, err := s.YoutubeService.GetVideoDetails(ctx, input)
	if err != nil {
		s.Logger.Error("Error al obtener detalles del video", zap.Error(err))
		return nil, fmt.Errorf("error al obtener detalles del video")
	}

	duration, err := parseCustomDuration(video.ContentDetails.Duration)
	if err != nil {
		s.Logger.Error("Error al analizar la duracion: ", zap.Error(err))
	}
	thumbnailURL := video.Snippet.Thumbnails.Default.Url

	song := &voice.Song{
		Type:         "youtube_provider",
		Title:        video.Snippet.Title,
		URL:          videoURL,
		Playable:     video.Snippet.LiveBroadcastContent != "live",
		ThumbnailURL: &thumbnailURL,
		Duration:     duration,
	}
	songs := []*voice.Song{song}

	s.Cache.Set(videoURL, songs)
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

// GetDCAData obtiene los datos de audio de una canción en formato DCA.
// Utiliza yt-dlp y ffmpeg para descargar el audio de YouTube y convertirlo al formato DCA esperado por Discord.
// Retorna un io.Reader que permite leer los datos de audio y un posible error.
func (s *YoutubeFetcher) GetDCAData(ctx context.Context, song *voice.Song) (io.Reader, error) {
	key := fmt.Sprintf("audio/%s.dca", song.Title)
	// Verificar si los datos de audio están en caché
	if cachedData, ok := s.audioCache.Get(song.URL); ok {
		return bytes.NewReader(cachedData), nil
	}

	// Verificar si el archivo está en S3
	exists, err := s.S3Uploader.FileExists(ctx, key)
	if err != nil {
		s.Logger.Error("Error al verificar la existencia del archivo en S3", zap.Error(err))
		return nil, fmt.Errorf("error al verificar la existencia del archivo en S3: %w", err)
	}

	var audioReader io.Reader

	if exists {
		// Descargar desde S3
		s.Logger.Info("Recuperando datos DCA de S3", zap.String("key", song.Title))
		audioReader, err = s.S3Uploader.DownloadDCA(ctx, key)
		if err != nil {
			s.Logger.Error("Error al descargar datos DCA desde S3", zap.Error(err))
			return nil, fmt.Errorf("error al descargar datos DCA desde S3: %w", err)
		}
	} else {
		// Crear un pipe para la transmisión progresiva de datos
		reader, writer := io.Pipe()

		go func() {
			defer writer.Close()

			// Buffer para almacenar los datos descargados y convertir en cache
			var buffer bytes.Buffer
			multiWriter := io.MultiWriter(writer, &buffer)

			if err := s.downloadAndStreamAudio(ctx, song, multiWriter); err != nil {
				s.Logger.Error("Error al descargar y transmitir audio", zap.Error(err))
				writer.CloseWithError(err)
				return
			}
			// Almacenar en cache
			s.audioCache.Set(song.URL, buffer.Bytes())

			// Subir a S3 si no existe
			if err := s.S3Uploader.UploadDCA(ctx, &buffer, key); err != nil {
				s.Logger.Error("Error al subir datos DCA a S3", zap.Error(err))
				// No devolvemos error aquí para no afectar la operación principal
			}
		}()

		audioReader = reader
	}

	return audioReader, nil
}

func (s *YoutubeFetcher) downloadAndStreamAudio(ctx context.Context, song *voice.Song, writer io.Writer) error {
	ytArgs := []string{"-f", "bestaudio[ext=m4a]", "--audio-quality", "0", "-o", "-", "--force-overwrites", "--http-chunk-size", "100K", "--username", "oauth2", "--password", "''", song.URL}
	ffmpegArgs := []string{"-i", "pipe:0", "-b:a", "192k", "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1"}

	cmd := s.CommandExecutor.ExecuteCommand(ctx, "sh", "-c", fmt.Sprintf("yt-dlp %s | ffmpeg %s | dca",
		strings.Join(ytArgs, " "),
		strings.Join(ffmpegArgs, " ")))

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error al crear el pipe de stdout: %w", err)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("error al crear el pipe de stderr: %w", err)
	}

	cmd.Stdout = writer

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		for scanner.Scan() {
			s.Logger.Info("yt-dlp stdout: %s", zap.String("Scanener", scanner.Text()))
		}
		if err := scanner.Err(); err != nil {
			s.Logger.Error("error leyendo stdout: %v", zap.Error(err))
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		for scanner.Scan() {
			// Registra la salida de error (errores y advertencias)
			s.Logger.Info("yt-dlp stderr: %s", zap.String("errorrr", scanner.Text()))
		}
		if err := scanner.Err(); err != nil {
			s.Logger.Error("error leyendo stderr: %v", zap.Error(err))
		}
	}()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("error al iniciar el comando: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("error al esperar el comando: %w", err)
	}

	return nil
}

func (s *YoutubeFetcher) SearchYouTubeVideoID(ctx context.Context, searchTerm string) (string, error) {
	videoID, err := s.YoutubeService.SearchVideoID(ctx, searchTerm)
	if err != nil {
		return "", fmt.Errorf("error al buscar el video en YouTube: %w", err)
	}
	return videoID, nil
}
