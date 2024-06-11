package fetcher

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/cache"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/voice"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/internal/metrics"
	"go.uber.org/zap"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	downloadBuffer = 1 * 1024 * 1024 // 1MB
)

// YoutubeFetcher es un tipo que interactúa con YouTube para obtener metadatos y datos de audio.
type YoutubeFetcher struct {
	Logger       logging.Logger
	Cache        cache.CacheManager
	CacheMetrics metrics.CacheMetrics
}

// NewYoutubeFetcher crea una nueva instancia de YoutubeFetcher con un logger predeterminado.
func NewYoutubeFetcher(logger logging.Logger, cache cache.CacheManager, cacheMetrics metrics.CacheMetrics) *YoutubeFetcher {
	return &YoutubeFetcher{
		Logger:       logger,
		Cache:        cache,
		CacheMetrics: cacheMetrics,
	}
}

// LookupSongs busca canciones en YouTube según el término de búsqueda proporcionado en input.
// Retorna una lista de objetos bot.Song que contienen metadatos de las canciones encontradas.
func (s *YoutubeFetcher) LookupSongs(ctx context.Context, input string) ([]*voice.Song, error) {
	cachedResult := s.Cache.Get(input)
	if cachedResult != nil {
		s.CacheMetrics.IncHits()
		return cachedResult, nil
	}
	// Define las columnas a imprimir por yt-dlp.
	ytDlpPrintColumns := []string{"title", "original_url", "is_live", "duration", "thumbnail", "thumbnails"}
	printColumns := strings.Join(ytDlpPrintColumns, ",")

	args := []string{"--print", printColumns, "--flat-playlist", "-U"}

	// Verifica si el input es una URL de YouTube o un término de búsqueda.
	if strings.HasPrefix(input, "https://") {
		args = append(args, input)
	} else {
		args = append(args, fmt.Sprintf("ytsearch:%s", input))
	}

	// Ejecuta yt-dlp con los argumentos especificados.
	ytCmd := exec.CommandContext(ctx, "yt-dlp", args...)

	ytOutBuf := &bytes.Buffer{}
	ytErrBuf := &bytes.Buffer{}
	ytCmd.Stdout = ytOutBuf
	ytCmd.Stderr = ytErrBuf

	// Ejecuta el comando yt-dlp y captura la salida.
	if err := ytCmd.Run(); err != nil {
		// Si hay un error, devuelve el mensaje de error capturado de yt-dlp
		return nil, fmt.Errorf("error al ejecutar el comando yt-dlp: %w, error_output: %s", err, ytErrBuf.String())
	}

	linesPerSong := len(ytDlpPrintColumns)
	ytOutLines := strings.Split(ytOutBuf.String(), "\n")
	songCount := len(ytOutLines) / linesPerSong

	songs := make([]*voice.Song, 0, songCount)
	for i := 0; i < songCount; i++ {
		duration, _ := strconv.ParseFloat(ytOutLines[linesPerSong*i+3], 32)

		var thumbnailURL *string = nil
		if ytOutLines[linesPerSong*i+4] != "NA" {
			thumbnailURL = &ytOutLines[linesPerSong*i+4]
		} else if ytOutLines[linesPerSong*i+5] != "NA" {
			thumbnail, err := getThumbnail(ytOutLines[linesPerSong*i+5])
			if err != nil {
				s.Logger.Error("error al obtener miniatura", zap.Error(err))
			}
			if thumbnail != nil {
				thumbnailURL = &thumbnail.URL
			}
		}

		// Crea un objeto Song con los metadatos obtenidos.
		song := &voice.Song{
			Type:         "yt-dlp",
			Title:        ytOutLines[linesPerSong*i],
			URL:          ytOutLines[linesPerSong*i+1],
			Playable:     ytOutLines[linesPerSong*i+2] == "False" || ytOutLines[3*i+2] == "NA",
			ThumbnailURL: thumbnailURL,
			Duration:     time.Second * time.Duration(duration),
		}
		if !song.Playable {
			continue
		}

		songs = append(songs, song)
	}
	s.Cache.Set(input, songs)
	s.CacheMetrics.IncMisses()
	s.CacheMetrics.SetCacheSize(float64(s.Cache.Size()))
	return songs, nil
}

// GetDCAData obtiene los datos de audio de una canción en formato DCA.
// Utiliza yt-dlp y ffmpeg para descargar el audio de YouTube y convertirlo al formato DCA esperado por Discord.
// Retorna un io.Reader que permite leer los datos de audio y un posible error.
func (s *YoutubeFetcher) GetDCAData(ctx context.Context, song *voice.Song) (io.Reader, error) {
	reader, writer := io.Pipe()

	go func(w io.WriteCloser) {
		defer func(w io.WriteCloser) {
			err := w.Close()
			if err != nil {
				s.Logger.Error("Error al cerrar el escritor: %v", zap.Error(err))
			}
		}(w)

		ytArgs := []string{"-f", "bestaudio[ext=m4a]", "--audio-quality", "0", "-o", "-", "--force-overwrites", "--http-chunk-size", "100K", "'" + song.URL + "'"}
		ffmpegArgs := []string{"-i", "pipe:0", "-b:a", "192k", "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1"}

		// Ejecuta una cadena de comandos para descargar el audio de YouTube y convertirlo a formato DCA.
		downloadCmd := exec.CommandContext(ctx,
			"sh", "-c", fmt.Sprintf("yt-dlp %s | ffmpeg %s | dca",
				strings.Join(ytArgs, " "),
				strings.Join(ffmpegArgs, " ")))

		bw := bufio.NewWriterSize(writer, downloadBuffer)
		downloadCmd.Stdout = bw
		// Ejecuta el comando y captura cualquier error.
		if err := downloadCmd.Run(); err != nil {
			s.Logger.Error("Error al ejecutar el comando: %v", zap.Error(err))
		}

		if err := bw.Flush(); err != nil {
			s.Logger.Error("mientras se limpia la tubería de datos DCA: %v", zap.Error(err))
		}
	}(writer)

	return reader, nil
}

// thumnail es una estructura que representa una miniatura de video.
type thumnail struct {
	URL        string `json:"url"`
	Preference int    `json:"preference"`
}

// getThumbnail analiza la cadena de entrada que representa información sobre miniaturas de video en formato JSON.
// Retorna la miniatura preferida como un objeto thumnail y un posible error.
func getThumbnail(thumnailsStr string) (*thumnail, error) {
	thumnailsStr = strings.ReplaceAll(thumnailsStr, "'", "\"")

	var thumbnails []thumnail
	if err := json.Unmarshal([]byte(thumnailsStr), &thumbnails); err != nil {
		return nil, err
	}

	if len(thumbnails) == 0 {
		return nil, nil
	}

	// Encuentra la miniatura con mayor preferencia.
	tn := &thumbnails[0]
	for i := range thumbnails {
		t := thumbnails[i]
		if t.Preference > tn.Preference {
			tn = &t
		}
	}

	return tn, nil
}
