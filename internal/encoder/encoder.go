package encoder

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"go.uber.org/zap"
	"io"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

var (
	AudioApplicationVoip     AudioApplication = "voip"
	AudioApplicationAudio    AudioApplication = "audio"
	AudioApplicationLowDelay AudioApplication = "lowdelay"
	StdEncodeOptions                          = &EncodeOptions{
		Volume:           256,
		Channels:         2,
		FrameRate:        48000,
		FrameDuration:    20,
		Bitrate:          64,
		Application:      AudioApplicationAudio,
		CompressionLevel: 10,
		PacketLoss:       1,
		BufferedFrames:   100,
		VBR:              true,
		StartTime:        0,
	}
)

// EncodeOptions es un conjunto de opciones para codificar dca
type (
	AudioApplication string

	EncodeOptions struct {
		Volume           int              // cambiar el volumen del audio (256 = normal)
		Channels         int              // canales de audio
		FrameRate        int              // frecuencia de muestreo del audio (por ej. 48000)
		FrameDuration    int              // duración del marco de audio, puede ser 20, 40 o 60 (ms)
		Bitrate          int              // tasa de bits de codificación de audio en kb/s, puede ser de 8 a 128
		PacketLoss       int              // porcentaje de pérdida de paquetes esperado
		RawOutput        bool             // Salida de opus en crudo (sin metadatos ni bytes mágicos)
		Application      AudioApplication // Aplicación de audio
		CoverFormat      string           // Formato con el que se codificará la carátula (por ej. "jpeg")
		CompressionLevel int              // Nivel de compresión, mayor es mejor calidad pero codificación más lenta (0 - 10)
		BufferedFrames   int              // Tamaño del búfer de cuadros
		VBR              bool             // Si se usa VBR o no (tasa de bits variable)
		Threads          int              // Número de hilos a utilizar, 0 para automático
		StartTime        int              // Tiempo de inicio de la secuencia de entrada en segundos
		AudioFilter      string
		Comment          string
	}

	Frame struct {
		data     []byte // datos del frame
		metadata bool   // si el cuadro contiene metadatos
	}

	EncodeStats struct {
		Size     int           // tamano del archivo codificado
		Duration time.Duration // duracion de la codificacion
		Bitrate  float32       // tasa de bits
		Speed    float32       // velocidad de procesamiento
	}

	// EncodeSession representa una sesión de codificación
	EncodeSession struct {
		sync.Mutex
		options      *EncodeOptions // Opciones de codificacion
		pipeReader   io.Reader      // lector para el pipe
		filePath     string         // ruta del archivo
		running      bool           // indica si la session esta en ejecuccion
		started      time.Time      // hora de inicio de la session
		frameChannel chan *Frame    // canal para transmitir los frames de audio
		process      *os.Process    // proceso de codificacion
		lastStats    *EncodeStats   // ultima estadisticas de codificacion
		lastFrame    int            // ultimo frame procesado
		err          error
		ffmpegOutput string // salida de ffmpeg
		logging      logging.ZapLogger
		// bufer para almacenar bytes no leidos (cuadros incompletos)
		// utilizado para impl io.Reader
		buf bytes.Buffer
	}
)

// PCMFrameLen calcula la longitud en cuadros PCM basada en las opciones de codificación.
// Devuelve el número de muestras de PCM para un cuadro dado.
func (e *EncodeOptions) PCMFrameLen() int {
	return 960 * e.Channels * (e.FrameDuration / 20)
}

// Validate valida las opciones de codificación para asegurarse de que están dentro de los límites permitidos.
// Devuelve un error si alguna opción es inválida.
func (opts *EncodeOptions) Validate() error {
	// Verifica que el volumen esté en el rango permitido.
	if opts.Volume < 0 || opts.Volume > 512 {
		return errors.New("Volumen fuera de los límites (0-512)")
	}

	// Verifica que la duración del cuadro sea una de las opciones válidas.
	if opts.FrameDuration != 20 && opts.FrameDuration != 40 && opts.FrameDuration != 60 {
		return errors.New("Duración del cuadro inválida")
	}

	// Verifica que el porcentaje de pérdida de paquetes esté en el rango permitido.
	if opts.PacketLoss < 0 || opts.PacketLoss > 100 {
		return errors.New("Porcentaje de pérdida de paquetes inválido")
	}

	// Verifica que la aplicación de audio sea una de las opciones válidas.
	if opts.Application != AudioApplicationAudio && opts.Application != AudioApplicationVoip && opts.Application != AudioApplicationLowDelay {
		return errors.New("Aplicación de audio inválida")
	}

	// Verifica que el nivel de compresión esté en el rango permitido.
	if opts.CompressionLevel < 0 || opts.CompressionLevel > 10 {
		return errors.New("Nivel de compresión fuera de los límites (0-10)")
	}

	// Verifica que el número de hilos no sea negativo.
	if opts.Threads < 0 {
		return errors.New("El número de hilos no puede ser menor que 0")
	}

	return nil
}

// EncodeMem crea una nueva sesión de codificación en memoria usando las opciones proporcionadas.
// Valida las opciones antes de iniciar la sesión. Devuelve la sesión de codificación o un error si las opciones son inválidas.
func EncodeMem(r io.Reader, options *EncodeOptions) (session *EncodeSession, err error) {
	err = options.Validate()
	if err != nil {
		return
	}

	session = &EncodeSession{
		options:      options,
		pipeReader:   r,
		frameChannel: make(chan *Frame, options.BufferedFrames),
	}
	go session.run()
	return
}

func (e *EncodeSession) run() {
	defer func() {
		e.Lock()
		e.running = false
		e.Unlock()
	}()
	e.Lock()
	e.running = true

	inFile := "pipe:0"
	if e.filePath != "" {
		inFile = e.filePath
	}

	if e.options == nil {
		e.options = StdEncodeOptions
	}

	vbrStr := "on"
	if !e.options.VBR {
		vbrStr = "off"
	}

	args := []string{
		"-stats",
		"-i", inFile,
		"-reconnect", "1",
		"-reconnect_at_eof", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "2",
		"-map", "0:a",
		"-acodec", "libopus",
		"-f", "ogg",
		"-vbr", vbrStr,
		"-compression_level", strconv.Itoa(e.options.CompressionLevel),
		"-af", fmt.Sprintf("volume=%.2f", e.options.Volume/100.0),
		"-ar", strconv.Itoa(e.options.FrameRate),
		"-ac", strconv.Itoa(e.options.Channels),
		"-b:a", strconv.Itoa(e.options.Bitrate * 1000),
		"-application", string(e.options.Application),
		"-frame_duration", strconv.Itoa(e.options.FrameDuration),
		"-packet_loss", strconv.Itoa(e.options.PacketLoss),
		"-threads", strconv.Itoa(e.options.Threads),
		"-ss", strconv.Itoa(e.options.StartTime),
	}

	if e.options.AudioFilter != "" {
		args = append(args, "-af", e.options.AudioFilter)
	}

	args = append(args, "pipe:1")

	ffmpeg := exec.Command("ffmpeg", args...)

	if e.pipeReader != nil {
		ffmpeg.Stdin = e.pipeReader
	}

	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		e.Unlock()
		e.logging.Error("StderrPipe Error", zap.Error(err))
		close(e.frameChannel)
		return
	}

	if !e.options.RawOutput {
		e.writeMetadataFrame()
	}

	err = ffmpeg.Start()
	if err != nil {
		e.Unlock()
		e.logging.Error("FFmpeg Start Error", zap.Error(err))
		close(e.frameChannel)
		return
	}

	e.started = time.Now()
	e.process = ffmpeg.Process
	e.Unlock()

	var wg sync.WaitGroup
	wg.Add(1)
	go e.readStderr(stdout, &wg)

	defer close(e.frameChannel)
	e.readStdout(stdout)
	wg.Wait()
	err = ffmpeg.Wait()
	if err != nil {
		if err.Error() != "signal: killed" {
			e.Lock()
			e.err = err
			e.Unlock()
		}
	}

}

func (e *EncodeSession) readStderr(stdout io.ReadCloser, s *sync.WaitGroup) {
	// TODO: implement
}

func (e *EncodeSession) readStdout(stdout io.ReadCloser) {
	// TODO: implement
}

func (e *EncodeSession) writeMetadataFrame() {
	// TODO: implement
}
