package encoder

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/internal/logging"
	"github.com/Tomas-vilte/GoMusicBot/internal/types"
	"go.uber.org/zap"
	"io"
	"mccoy.space/g/ogg"
	"os"
	"os/exec"
	"strconv"
	"strings"
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
func (e *EncodeOptions) Validate() error {
	// Verifica que el volumen esté en el rango permitido.
	if e.Volume < 0 || e.Volume > 512 {
		return ErrInvalidVolume
	}

	// Verifica que la duración del cuadro sea una de las opciones válidas.
	if e.FrameDuration != 20 && e.FrameDuration != 40 && e.FrameDuration != 60 {
		return ErrInvalidFrameDuration
	}

	// Verifica que el porcentaje de pérdida de paquetes esté en el rango permitido.
	if e.PacketLoss < 0 || e.PacketLoss > 100 {
		return ErrInvalidPacketLoss
	}

	// Verifica que la aplicación de audio sea una de las opciones válidas.
	if e.Application != AudioApplicationAudio && e.Application != AudioApplicationVoip && e.Application != AudioApplicationLowDelay {
		return ErrInvalidAudioApplication
	}

	// Verifica que el nivel de compresión esté en el rango permitido.
	if e.CompressionLevel < 0 || e.CompressionLevel > 10 {
		return ErrInvalidCompressionLevel
	}

	// Verifica que el número de hilos no sea negativo.
	if e.Threads < 0 {
		return ErrInvalidThreads
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

func EncodeFile(path string, options *EncodeOptions) (session *EncodeSession, err error) {
	logger, err := logging.NewZapLogger(false)
	if err != nil {
		return nil, err
	}
	err = options.Validate()
	if err != nil {
		return
	}
	session = &EncodeSession{
		options:      options,
		filePath:     path,
		frameChannel: make(chan *Frame, options.BufferedFrames),
		logging:      *logger,
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
		"-af", fmt.Sprintf("volume=%.2f", float64(e.options.Volume)/100.0),
		"-ar", strconv.Itoa(e.options.FrameRate),
		"-ac", strconv.Itoa(e.options.Channels),
		"-b:a", strconv.Itoa(e.options.Bitrate * 1000),
		"-application", string(e.options.Application),
		"-frame_duration", strconv.Itoa(e.options.FrameDuration),
		"-packet_loss", strconv.Itoa(e.options.PacketLoss),
		"-threads", strconv.Itoa(e.options.Threads),
		"-ss", strconv.Itoa(e.options.StartTime),
	}

	args = append(args, "pipe:1")

	ffmpeg := exec.Command("ffmpeg", args...)

	e.logging.Debug("Ejecutando ffmpeg",
		zap.Strings("args", ffmpeg.Args),
		zap.String("input_file", inFile))

	if e.pipeReader != nil {
		ffmpeg.Stdin = e.pipeReader
	}

	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		e.logging.Error("Error al obtener stdout de ffmpeg", zap.Error(err))
		e.setError(ErrFailedToReadStdout)
		close(e.frameChannel)
		return
	}

	stderr, err := ffmpeg.StderrPipe()
	if err != nil {
		e.logging.Error("Error al obtener stderr de ffmpeg", zap.Error(err))
		e.setError(ErrFailedToReadStderr)
		close(e.frameChannel)
		return
	}

	if !e.options.RawOutput {
		e.writeMetadataFrame()
	}

	err = ffmpeg.Start()
	if err != nil {
		e.logging.Error("Error al iniciar ffmpeg", zap.Error(err))
		e.setError(ErrFailedToStartFFMPEG)
		close(e.frameChannel)
		return
	}

	e.started = time.Now()

	e.process = ffmpeg.Process
	e.Unlock()

	var wg sync.WaitGroup
	wg.Add(1)
	go e.readStderr(stderr, &wg)

	defer close(e.frameChannel)
	e.readStdout(stdout)
	wg.Wait()
	err = ffmpeg.Wait()
	if err != nil {
		if err.Error() != "signal: killed" {
			e.logging.Error("Error al esperar a ffmpeg", zap.Error(err))
			e.setError(err)
		}
	}
}

func (e *EncodeSession) setError(err error) {
	e.Lock()
	defer e.Unlock()
	if e.err == nil {
		e.err = err
	}
}

func (e *EncodeSession) readStderr(stderr io.ReadCloser, wg *sync.WaitGroup) {
	defer wg.Done()

	bufReader := bufio.NewReader(stderr)
	var outBuf bytes.Buffer
	for {
		r, _, err := bufReader.ReadRune()
		if err != nil {
			if err != io.EOF {
				e.logging.Error("Error leyendo stderr", zap.Error(err))

			}
			break
		}

		switch r {
		case '\r':
			if outBuf.Len() > 0 {
				e.handleStderrLine(outBuf.String())
				outBuf.Reset()
			}
		case '\n':
			e.Lock()
			e.ffmpegOutput += outBuf.String() + "\n"
			e.Unlock()
			outBuf.Reset()
		default:
			outBuf.WriteRune(r)
		}
	}
}

func (e *EncodeSession) handleStderrLine(line string) {
	if strings.Index(line, "size=") != 0 {
		return
	}

	var size int

	var timeH int
	var timeM int
	var timeS float32

	var bitrate float32
	var speed float32

	_, err := fmt.Sscanf(line, "size=%dkB time=%d:%d:%f bitrate=%fkbits/s speed=%fx", &size, &timeH, &timeM, &timeS, &bitrate, &speed)
	if err != nil {
		e.logging.Error("Error al analizar línea de stderr", zap.Error(err))
	}

	dur := time.Duration(timeH) * time.Hour
	dur += time.Duration(timeM) * time.Minute
	dur += time.Duration(timeS) * time.Second

	stats := &EncodeStats{
		Size:     size,
		Duration: dur,
		Bitrate:  bitrate,
		Speed:    speed,
	}

	e.Lock()
	e.lastStats = stats
	e.Unlock()
}

func (e *EncodeSession) readStdout(stdout io.ReadCloser) {
	decoder := NewPacketDecoder(ogg.NewDecoder(stdout))

	skipPackets := 2
	for {
		packet, _, err := decoder.Decode()
		if skipPackets > 0 {
			skipPackets--
			continue
		}
		if err != nil {
			if err != io.EOF {
				e.logging.Error("Error al leer stdout", zap.Error(err))
			}
			break
		}
		err = e.writeOpusFrame(packet)
		if err != nil {
			e.logging.Error("Error escribir opus frame", zap.Error(err))
			break
		}
	}
}

func (e *EncodeSession) writeOpusFrame(opusFrame []byte) error {
	var dcaBuf bytes.Buffer

	err := binary.Write(&dcaBuf, binary.LittleEndian, int16(len(opusFrame)))
	if err != nil {
		e.logging.Error("Error al escribir datos de frame DCA", zap.Error(err))
		return err
	}

	_, err = dcaBuf.Write(opusFrame)
	if err != nil {
		e.logging.Error("Error al escribir frame Opus", zap.Error(err))
		return err
	}

	e.frameChannel <- &Frame{dcaBuf.Bytes(), false}

	e.Lock()
	e.lastFrame++
	e.Unlock()

	return nil
}

func (e *EncodeSession) writeMetadataFrame() {
	metadata := types.Metadata{
		Opus: &types.OpusMetadata{
			Bitrate:     e.options.Bitrate * 1000,
			SampleRate:  e.options.FrameRate,
			Application: string(e.options.Application),
			FrameSize:   e.options.PCMFrameLen(),
			Channels:    e.options.Channels,
			VBR:         e.options.VBR,
		},
		Origin: &types.OriginMetadata{
			Source:   "file",
			Channels: e.options.Channels,
			Bitrate:  e.options.Bitrate * 1000,
			Encoding: "Opus",
		},
	}

	var buf bytes.Buffer
	buf.Write([]byte(fmt.Sprintf("DCA%d", 1)))

	jsonData, err := json.Marshal(metadata)
	if err != nil {
		e.logging.Error("Error al codificar metadatos en JSON", zap.Error(err))
		return
	}

	jsonLen := int32(len(jsonData))
	err = binary.Write(&buf, binary.LittleEndian, &jsonLen)
	if err != nil {
		e.logging.Error("Error al escribir longitud de JSON", zap.Error(err))
		return
	}

	buf.Write(jsonData)
	e.frameChannel <- &Frame{buf.Bytes(), true}
}

func (e *EncodeSession) ReadFrame() (frame []byte, err error) {
	f := <-e.frameChannel
	if f == nil {
		return nil, io.EOF
	}

	return f.data, nil
}

func (e *EncodeSession) Read(p []byte) (n int, err error) {
	if e.buf.Len() >= len(p) {
		return e.buf.Read(p)
	}

	for e.buf.Len() < len(p) {
		f, err := e.ReadFrame()
		if err != nil {
			if err == io.EOF {
				break
			}
			e.logging.Error("Error al leer frame", zap.Error(err))
			return 0, err
		}
		e.buf.Write(f)
	}
	return e.buf.Read(p)
}

func (e *EncodeSession) FFMPEGMessages() string {
	e.Lock()
	output := e.ffmpegOutput
	e.Unlock()
	return output
}

func (e *EncodeSession) Stop() error {
	e.Lock()
	defer e.Unlock()
	if !e.running || e.process == nil {
		return errors.New("la session no esta corriendo")
	}
	if err := e.process.Kill(); err != nil {
		e.logging.Error("Error al detener el proceso de codificación", zap.Error(err))
		return err
	}
	e.running = false
	return nil
}
