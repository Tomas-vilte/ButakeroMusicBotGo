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
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	downloadBuffer = 100 * 1024 // 100 KiB
)

// YoutubeFetcher es una implementación de Fetcher que utiliza yt-dlp y ffmpeg.
type YoutubeFetcher struct{}

func NewYoutubeFetcher() *YoutubeFetcher {
	return &YoutubeFetcher{}
}

// LookupSongs busca canciones en YouTube y devuelve metadatos de canciones.
func (yf *YoutubeFetcher) LookupSongs(ctx context.Context, query string) ([]*bot.Song, error) {
	// Columnas a imprimir de la salida de yt-dlp
	ytDlpPrintColumns := []string{"title", "original_url", "is_live", "duration", "thumbnail", "thumbnails"}
	printColumns := strings.Join(ytDlpPrintColumns, ",")

	// Argumentos para yt-dlp
	args := []string{"--print", printColumns, "--flat-playlist", "-U"}

	// Si la consulta es una URL, la añadimos a los argumentos, de lo contrario realizamos una búsqueda
	if strings.HasPrefix(query, "https://") {
		args = append(args, query)
	} else {
		args = append(args, fmt.Sprintf("ytsearch:%s", query))
	}

	// Ejecutar yt-dlp con los argumentos especificados
	ytCmd := exec.CommandContext(ctx, "yt-dlp", args...)

	// Capturar la salida estándar de yt-dlp
	ytOutBuf := &bytes.Buffer{}
	ytCmd.Stdout = ytOutBuf

	// Ejecutar el comando yt-dlp
	if err := ytCmd.Run(); err != nil {
		return nil, fmt.Errorf("al ejecutar el comando yt-dlp para obtener metadatos: %w", err)
	}

	// Calcular el número de líneas por canción
	linesPerSong := len(ytDlpPrintColumns)
	ytOutLines := strings.Split(ytOutBuf.String(), "\n")
	songCount := len(ytOutLines) / linesPerSong

	// Crear un slice para almacenar las canciones
	songs := make([]*bot.Song, 0, songCount)
	for i := 0; i < songCount; i++ {
		// Parsear la duración de la canción a float64
		duration, _ := strconv.ParseFloat(ytOutLines[linesPerSong*i+3], 32)

		// Determinar la URL de la miniatura de la canción
		var thumbnailURL *string = nil
		if ytOutLines[linesPerSong*i+4] != "NA" {
			thumbnailURL = &ytOutLines[linesPerSong*i+4]
		} else if ytOutLines[linesPerSong*i+5] != "NA" {
			thumbnail, err := getThumbnail(ytOutLines[linesPerSong*i+5])
			if err != nil {
				log.Println("Error al obtener la miniatura:", err)
			}
			if thumbnail != nil {
				thumbnailURL = &thumbnail.URL
			}
		}

		// Crear un objeto Song con la información obtenida
		song := &bot.Song{
			Type:         "youtube",
			Title:        ytOutLines[linesPerSong*i],
			URL:          ytOutLines[linesPerSong*i+1],
			Playable:     ytOutLines[linesPerSong*i+2] == "False" || ytOutLines[linesPerSong*i+2] == "NA",
			ThumbnailURL: thumbnailURL,
			Duration:     time.Second * time.Duration(duration),
		}
		// Si la canción no es reproducible, saltarla
		if !song.Playable {
			continue
		}

		// Agregar la canción al slice
		songs = append(songs, song)
	}

	return songs, nil
}

// GetDCAData obtiene datos de audio DCA para una canción de YouTube.
func (y *YoutubeFetcher) GetDCAData(ctx context.Context, song *bot.Song) (io.Reader, error) {
	// Crear una tubería para leer y escribir datos
	reader, writer := io.Pipe()

	go func(w io.WriteCloser) {
		defer w.Close()

		// Argumentos para yt-dlp y ffmpeg
		ytArgs := []string{"-U", "-x", "-o", "-", "--force-overwrites", "--http-chunk-size", "100K", "'" + song.URL + "'"}
		ffmpegArgs := []string{"i", "pipe:0"}

		// Si existe una posición inicial especificada, añadirla a los argumentos de ffmpeg
		if song.StartPosition > 0 {
			ffmpegArgs = append(ffmpegArgs, "-ss", song.StartPosition.String())
		}
		ffmpegArgs = append(ffmpegArgs, "-f", "s16le", "-ar", "48000", "-ac", "2", "pipe:1")

		// comando para descargar y convertirlo en audio
		downloadCmd := exec.CommandContext(ctx,
			"sh", "-c", fmt.Sprintf("yt-dlp %s | ffmpeg %s | dca",
				strings.Join(ytArgs, " "),
				strings.Join(ffmpegArgs, " ")))

		// buffer de escritura con tamaño específico
		bw := bufio.NewWriterSize(writer, downloadBuffer)
		downloadCmd.Stdout = bw

		// Ejecutar el comando
		if err := downloadCmd.Run(); err != nil {
			log.Printf("Error al ejecutar el comando para obtener datos DCA: %v", err)
		}

		// Vaciar el buffer de escritura
		if err := bw.Flush(); err != nil {
			log.Printf("Error al vaciar el buffer de escritura de datos DCA: %v", err)
		}
	}(writer)

	// Devolver el lector para los datos DCA y cualquier error
	return reader, nil
}

type thumnail struct {
	URL        string `json:"url"`
	Preference int    `json:"preference"`
}

// getThumbnail obtiene la URL de la miniatura de una cadena JSON.
func getThumbnail(thumnailsStr string) (*thumnail, error) {
	// Reemplazar comillas simples por comillas dobles para analizar JSON
	thumnailsStr = strings.ReplaceAll(thumnailsStr, "'", "\"")

	// Crear una variable para almacenar las miniaturas
	var thumbnails []thumnail
	// Parsear la cadena JSON en la lista de miniaturas
	if err := json.Unmarshal([]byte(thumnailsStr), &thumbnails); err != nil {
		return nil, err
	}

	// Si no hay miniaturas, devolver nil
	if len(thumbnails) == 0 {
		return nil, nil
	}

	// Encontrar la miniatura con la mayor preferencia
	tn := &thumbnails[0]
	for i := range thumbnails {
		t := thumbnails[i]
		if t.Preference > tn.Preference {
			tn = &t
		}
	}

	return tn, nil
}
