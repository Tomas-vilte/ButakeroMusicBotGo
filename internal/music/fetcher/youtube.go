package fetcher

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/discord/bot"
	"io"
	"log"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	downloadBuffer = 100 * 1024 // 100 KiB
)

// YoutubeFetcher es un tipo que interactúa con YouTube para obtener metadatos y datos de audio.
type YoutubeFetcher struct {
	Logger *slog.Logger
}

// NewYoutubeFetcher crea una nueva instancia de YoutubeFetcher con un logger predeterminado.
func NewYoutubeFetcher() *YoutubeFetcher {
	return &YoutubeFetcher{
		Logger: slog.Default(),
	}
}

// LookupSongs busca canciones en YouTube según el término de búsqueda proporcionado en input.
// Retorna una lista de objetos bot.Song que contienen metadatos de las canciones encontradas.
func (s *YoutubeFetcher) LookupSongs(ctx context.Context, input string) ([]*bot.Song, error) {
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
	ytCmd.Stdout = ytOutBuf

	// Ejecuta el comando yt-dlp y captura la salida.
	if err := ytCmd.Run(); err != nil {
		return nil, fmt.Errorf("al ejecutar el comando yt-dlp para obtener metadatos: %w", err)
	}

	linesPerSong := len(ytDlpPrintColumns)
	ytOutLines := strings.Split(ytOutBuf.String(), "\n")
	songCount := len(ytOutLines) / linesPerSong

	songs := make([]*bot.Song, 0, songCount)
	for i := 0; i < songCount; i++ {
		duration, _ := strconv.ParseFloat(ytOutLines[linesPerSong*i+3], 32)

		var thumbnailURL *string = nil
		if ytOutLines[linesPerSong*i+4] != "NA" {
			thumbnailURL = &ytOutLines[linesPerSong*i+4]
		} else if ytOutLines[linesPerSong*i+5] != "NA" {
			thumbnail, err := getThumbnail(ytOutLines[linesPerSong*i+5])
			if err != nil {
				s.Logger.Error("error al obtener miniatura", "error", err)
			}
			if thumbnail != nil {
				thumbnailURL = &thumbnail.URL
			}
		}

		// Crea un objeto Song con los metadatos obtenidos.
		song := &bot.Song{
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

	return songs, nil
}

// GetDCAData obtiene los datos de audio de una canción en formato DCA.
// Utiliza yt-dlp y ffmpeg para descargar el audio de YouTube y convertirlo al formato DCA esperado por Discord.
// Retorna un io.Reader que permite leer los datos de audio y un posible error.
func (s *YoutubeFetcher) GetDCAData(ctx context.Context, song *bot.Song) (io.Reader, error) {
	reader, writer := io.Pipe()

	go func(w io.WriteCloser) {
		defer w.Close()

		ytArgs := []string{"-U", "-x", "-o", "-", "--force-overwrites", "--http-chunk-size", "100K", "'" + song.URL + "'"}
		ffmpegArgs := []string{"-i", "pipe:0"}

		if song.StartPosition > 0 {
			ffmpegArgs = append(ffmpegArgs, "-ss", song.StartPosition.String())
		}
		ffmpegArgs = append(ffmpegArgs, "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1")

		// Ejecuta una cadena de comandos para descargar el audio de YouTube y convertirlo a formato DCA.
		downloadCmd := exec.CommandContext(ctx,
			"sh", "-c", fmt.Sprintf("yt-dlp %s | ffmpeg %s | dca",
				strings.Join(ytArgs, " "),
				strings.Join(ffmpegArgs, " ")))

		bw := bufio.NewWriterSize(writer, downloadBuffer)
		downloadCmd.Stdout = bw

		// Ejecuta el comando y captura cualquier error.
		if err := downloadCmd.Run(); err != nil {
			log.Printf("al ejecutar el tubo para obtener los datos DCA: %v", err)
		}

		if err := bw.Flush(); err != nil {
			log.Printf("al vaciar el tubo de datos DCA: %v", err)
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
