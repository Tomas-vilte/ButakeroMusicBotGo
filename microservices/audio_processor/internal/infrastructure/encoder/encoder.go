package encoder

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"io"
	"mccoy.space/g/ogg"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	AudioApplicationVoip     AudioApplication = "voip"     // Aplicación de audio para voz sobre IP (VoIP)
	AudioApplicationAudio    AudioApplication = "audio"    // Aplicación de audio general
	AudioApplicationLowDelay AudioApplication = "lowdelay" // Aplicación de audio con baja latencia

	// StdEncodeOptions Opciones predeterminadas para la codificación de audio.
	StdEncodeOptions = &EncodeOptions{
		Volume:           256,                   // Nivel de volumen (256 es el valor normal)
		Channels:         2,                     // Número de canales de audio (por ej. 2 para estéreo)
		FrameRate:        48000,                 // Frecuencia de muestreo del audio en Hz (por ej. 48000 Hz)
		FrameDuration:    20,                    // Duración del marco de audio en ms (puede ser 20, 40 o 60 ms)
		Bitrate:          96,                    // Tasa de bits en kb/s (por ej. 64 kb/s)
		Application:      AudioApplicationAudio, // Aplicación de audio a usar
		CompressionLevel: 10,                    // Nivel de compresión (0 a 10, donde 10 es la máxima compresión y menor velocidad de codificación)
		PacketLoss:       1,                     // Porcentaje de pérdida de paquetes esperado
		BufferedFrames:   100,                   // Tamaño del búfer de cuadros
		VBR:              true,                  // Si se usa VBR (tasa de bits variable) o no
		StartTime:        0,                     // Tiempo de inicio de la secuencia de entrada en segundos
	}
)

type (
	AudioEncoder interface {
		Encode(ctx context.Context, r io.Reader, options *EncodeOptions) (EncodeSession, error)
	}

	EncodeSession interface {
		ReadFrame() ([]byte, error)
		Read(p []byte) (n int, err error)
		Stop() error
		FFMPEGMessages() string
		Cleanup()
	}

	FFmpegEncoder struct {
		log logger.Logger
	}
)

func NewFFmpegEncoder(log logger.Logger) *FFmpegEncoder {
	return &FFmpegEncoder{
		log: log,
	}
}

// PCMFrameLen calcula la longitud en cuadros PCM basada en las opciones de codificación.
// Devuelve el número de muestras de PCM para un cuadro dado.
func (opts *EncodeOptions) PCMFrameLen() int {
	return 960 * opts.Channels * (opts.FrameDuration / 20)
}

// Validate valida las opciones de codificación para asegurarse de que están dentro de los límites permitidos.
// Devuelve un error si alguna opción es inválida.
func (opts *EncodeOptions) Validate() error {
	if opts.Volume < 0 || opts.Volume > 512 {
		return ErrInvalidVolume
	}

	if opts.FrameDuration != 20 && opts.FrameDuration != 40 && opts.FrameDuration != 60 {
		return ErrInvalidFrameDuration
	}

	if opts.PacketLoss < 0 || opts.PacketLoss > 100 {
		return ErrInvalidPacketLoss
	}

	if opts.Application != AudioApplicationAudio && opts.Application != AudioApplicationVoip && opts.Application != AudioApplicationLowDelay {
		return ErrInvalidAudioApplication
	}

	if opts.CompressionLevel < 0 || opts.CompressionLevel > 10 {
		return ErrInvalidCompressionLevel
	}

	if opts.Threads < 0 {
		return ErrInvalidThreads
	}

	return nil
}

// Encode crea una nueva sesión de codificación en memoria usando las opciones proporcionadas.
// Valida las opciones antes de iniciar la sesión. Devuelve la sesión de codificación o un error si las opciones son inválidas.
func (f *FFmpegEncoder) Encode(ctx context.Context, r io.Reader, options *EncodeOptions) (EncodeSession, error) {
	if err := options.Validate(); err != nil {
		return nil, err
	}

	session := &encodeSessionimpl{
		options:      options,
		pipeReader:   r,
		frameChannel: make(chan *Frame, options.BufferedFrames),
		log:          f.log,
	}
	go session.run(ctx)
	return session, nil
}

// run ejecuta el proceso de codificación de audio utilizando ffmpeg.
// Este método se ejecuta en una goroutine y gestiona la ejecución del comando ffmpeg,
// la lectura de su salida y la escritura de metadatos si es necesario.
//
// Parámetros:
// - ctx: Contexto que permite cancelar el proceso de codificación.
//
// Detalles del funcionamiento:
// - Inicializa y configura el comando ffmpeg con los argumentos adecuados.
// - Maneja la entrada y salida del proceso ffmpeg.
// - Escribe metadatos en el archivo de salida si las opciones lo requieren.
// - Lee la salida estándar y los errores del proceso ffmpeg.
// - Utiliza un grupo de espera para sincronizar la lectura de stderr.
// - Maneja el estado de la sesión y actualiza los logs con información relevante.
func (e *encodeSessionimpl) run(ctx context.Context) {
	// Reset running state
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
		// Lit af
		args = append(args, "-af", e.options.AudioFilter)
	}

	args = append(args, "pipe:1")

	ffmpeg := exec.CommandContext(ctx, "ffmpeg", args...)

	if e.pipeReader != nil {
		ffmpeg.Stdin = e.pipeReader
	}

	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		e.Unlock()
		e.log.Error("Error al obtener stdout de ffmpeg", zap.Error(err))
		close(e.frameChannel)
		return
	}

	stderr, err := ffmpeg.StderrPipe()
	if err != nil {
		e.Unlock()
		e.log.Error("Error al obtener stderr de ffmpeg", zap.Error(err))
		close(e.frameChannel)
		return
	}

	if !e.options.RawOutput {
		e.writeMetadataFrame()
	}

	err = ffmpeg.Start()
	if err != nil {
		e.Unlock()
		e.log.Error("Error al iniciar ffmpeg", zap.Error(err))
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
			e.Lock()
			e.err = err
			e.Unlock()
		}
	}
}

func (e *encodeSessionimpl) writeMetadataFrame() {
	// Crea los metadatos de la codificación Opus y el origen del archivo.
	metadata := Metadata{
		Opus: &OpusMetadata{
			Bitrate:     e.options.Bitrate * 1000,
			SampleRate:  e.options.FrameRate,
			Application: string(e.options.Application),
			FrameSize:   e.options.PCMFrameLen(),
			Channels:    e.options.Channels,
			VBR:         e.options.VBR,
		},
		Origin: &OriginMetadata{
			Source:   "file",
			Channels: e.options.Channels,
			Bitrate:  e.options.Bitrate * 1000,
			Encoding: "Opus",
		},
	}

	var buf bytes.Buffer
	// Escribe el encabezado "DCA1" en el búfer.
	buf.Write([]byte(fmt.Sprintf("DCA%d", 1)))

	// Serializa los metadatos a formato JSON.
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		e.log.Error("Error al codificar metadatos en JSON", zap.Error(err))
		return
	}

	// Escribe la longitud del JSON en formato Little Endian.
	if err := binary.Write(&buf, binary.LittleEndian, int32(len(jsonData))); err != nil {
		e.log.Error("Error al escribir longitud de JSON", zap.Error(err))
		return
	}

	// Escribe el JSON serializado en el búfer.
	buf.Write(jsonData)
	// Envía el búfer con el frame de metadatos al canal de frames.
	e.frameChannel <- &Frame{buf.Bytes(), true}
}

func (e *encodeSessionimpl) readStderr(stderr io.ReadCloser, wg *sync.WaitGroup) {
	defer wg.Done()

	bufReader := bufio.NewReader(stderr)
	var outBuf bytes.Buffer
	for {
		r, _, err := bufReader.ReadRune()
		if err != nil {
			if err != io.EOF {
				e.log.Error("Error leyendo stderr", zap.Error(err))
			}
			break
		}

		switch r {
		case '\r':
			// Stats line
			if outBuf.Len() > 0 {
				e.handleStderrLine(outBuf.String())
				outBuf.Reset()
			}
		case '\n':
			// Message
			e.Lock()
			e.ffmpegOutput += outBuf.String() + "\n"
			e.Unlock()
			outBuf.Reset()
		default:
			outBuf.WriteRune(r)
		}
	}
}

// handleStderrLine procesa una línea de la salida de error (stderr) del proceso ffmpeg,
// extrayendo información relevante sobre el progreso de la codificación.
//
// Parámetros:
// - line: La línea de texto leída de stderr que contiene información sobre el tamaño, el tiempo, la tasa de bits y la velocidad.
//
// Detalles del funcionamiento:
// - Verifica si la línea comienza con "size=", que indica que contiene información útil.
// - Extrae y analiza los valores de tamaño, tiempo, tasa de bits y velocidad utilizando fmt.Sscanf.
// - Calcula la duración total en base a las horas, minutos y segundos extraídos.
// - Crea una instancia de EncodeStats con la información extraída y la almacena en el campo lastStats de la sesión.
// - Utiliza un mutex para garantizar que el acceso a lastStats sea seguro en un entorno concurrente.
func (e *encodeSessionimpl) handleStderrLine(line string) {
	if strings.Index(line, "size=") != 0 {
		return // no hay info
	}

	fields := strings.Fields(line)
	stats := &EncodeStats{}

	for _, field := range fields {
		switch {
		case strings.HasPrefix(field, "size="):
			size, _ := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(field, "size="), "kB"))
			stats.Size = size
		case strings.HasPrefix(field, "time="):
			timeStr := strings.TrimPrefix(field, "time=")
			timeParts := strings.Split(timeStr, ":")
			if len(timeParts) == 3 {
				h, _ := strconv.Atoi(timeParts[0])
				m, _ := strconv.Atoi(timeParts[1])
				s, _ := strconv.ParseFloat(timeParts[2], 32)
				stats.Duration = time.Duration(h)*time.Hour +
					time.Duration(m)*time.Minute +
					time.Duration(s*float64(time.Second))
			}
		case strings.HasPrefix(field, "bitrate="):
			bitrate, _ := strconv.ParseFloat(strings.TrimSuffix(strings.TrimPrefix(field, "bitrate="), "kbits/s"), 32)
			stats.Bitrate = float32(bitrate)
		case strings.HasPrefix(field, "speed="):
			speed, _ := strconv.ParseFloat(strings.TrimSuffix(strings.TrimPrefix(field, "speed="), "x"), 32)
			stats.Speed = float32(speed)
		}
	}

	e.Lock()
	e.lastStats = stats
	e.Unlock()
}

// readStdout lee la salida estándar (stdout) del proceso ffmpeg y procesa los paquetes de audio en formato Opus.
//
// Parámetros:
// - stdout: Un io.ReadCloser que proporciona la salida estándar del proceso ffmpeg.
//
// Detalles del funcionamiento:
// - Crea un decodificador de paquetes usando la salida estándar de ffmpeg y un decodificador OGG.
// - Omite los primeros dos paquetes, que generalmente son innecesarios para el procesamiento.
// - En un bucle continuo, decodifica los paquetes de audio desde stdout.
// - Si ocurre un error durante la decodificación, se registra el error y se detiene la lectura, excepto en el caso de EOF.
// - Escribe los paquetes decodificados en el formato Opus utilizando el método `writeOpusFrame`.
// - Registra errores si ocurren durante la escritura de los frames Opus y detiene el procesamiento si es necesario.
func (e *encodeSessionimpl) readStdout(stdout io.ReadCloser) {
	decoder := NewPacketDecoder(ogg.NewDecoder(stdout))

	// los primeros 2 paquetes son metadatos de ogg opus
	skipPackets := 2
	for {
		// Recupera paquete
		packet, _, err := decoder.Decode()
		if skipPackets > 0 {
			skipPackets--
			continue
		}
		if err != nil {
			if err != io.EOF {
				e.log.Error("Error al leer stdout", zap.Error(err))
			}
			break
		}

		err = e.writeOpusFrame(packet)
		if err != nil {
			e.log.Error("Error escribir opus frame", zap.Error(err))
			break
		}
	}
}

// writeOpusFrame escribe un frame de audio en formato Opus en el canal de frames y actualiza el contador de frames.
//
// Parámetros:
// - opusFrame: Un slice de bytes que representa el frame de audio en formato Opus.
//
// Retorna:
// - Un error si ocurre algún problema al escribir el frame; nil si la operación es exitosa.
//
// Detalles del funcionamiento:
// - Crea un búfer para almacenar los datos en el formato DCA (Dolby Coherent Audio).
// - Escribe el tamaño del frame Opus como un entero de 16 bits en el búfer en formato Little Endian.
// - Escribe el frame Opus en el búfer.
// - Envía el búfer con los datos del frame a través del canal de frames.
// - Incrementa el contador de frames procesados de manera segura utilizando un mutex.
// - Registra errores si ocurren durante la escritura de los datos del frame y devuelve el error.
func (e *encodeSessionimpl) writeOpusFrame(opusFrame []byte) error {
	var dcaBuf bytes.Buffer

	err := binary.Write(&dcaBuf, binary.LittleEndian, int16(len(opusFrame)))
	if err != nil {
		return err
	}

	_, err = dcaBuf.Write(opusFrame)
	if err != nil {
		return err
	}

	e.frameChannel <- &Frame{dcaBuf.Bytes(), false}

	e.Lock()
	e.lastFrame++
	e.Unlock()

	return nil
}

// Stop detiene la sesión de codificación si está en ejecución.
//
// Retorna:
// - Un error si ocurre algún problema al intentar detener el proceso; nil si la operación es exitosa.
//
// Detalles del funcionamiento:
//   - Adquiere un bloqueo en el mutex para asegurar el acceso seguro a los atributos de la sesión.
//   - Verifica si la sesión está en ejecución y si el proceso de codificación está activo.
//     Si no es así, retorna un error indicando que la sesión no está corriendo.
//   - Intenta detener el proceso de codificación llamando a `Kill` en el proceso.
//     Si ocurre un error durante esta operación, lo registra y lo retorna.
//   - Actualiza el estado de la sesión para indicar que ya no está en ejecución.
//   - Libera el bloqueo en el mutex antes de retornar.
func (e *encodeSessionimpl) Stop() error {
	e.Lock()
	defer e.Unlock()

	if !e.running || e.process == nil {
		return errors.New("la session no esta corriendo")
	}

	if err := e.process.Kill(); err != nil {
		e.log.Error("Error al detener el proceso de codificación", zap.Error(err))
		return err
	}
	e.running = false
	return nil
}

// ReadFrame lee un frame de audio del canal de frames y lo devuelve como un slice de bytes.
//
// Retorna:
// - Un slice de bytes que contiene los datos del frame leído del canal de frames.
// - Un error que es io.EOF si el canal está cerrado y no hay más frames disponibles.
//
// Detalles del funcionamiento:
//   - Lee un frame del canal de frames. Si el frame es nil, indica que el canal ha sido cerrado y no hay más datos disponibles,
//     en cuyo caso retorna io.EOF.
//   - Devuelve los datos del frame como un slice de bytes si la operación es exitosa.
func (e *encodeSessionimpl) ReadFrame() (frame []byte, err error) {
	f := <-e.frameChannel
	if f == nil {
		return nil, io.EOF
	}

	return f.data, nil
}

// Running devuelve true si se está ejecutando
func (e *encodeSessionimpl) Running() (running bool) {
	e.Lock()
	running = e.running
	e.Unlock()
	return
}

// Stats devuelve estadísticas de ffmpeg. NOTA: no se trata de estadísticas de reproducción sino de transcodificación.
// Para saber qué tan avanzado estás en la reproducción
// tenes que realizar un seguimiento del número de fotogramas enviados a Discord vos mismo
func (e *encodeSessionimpl) Stats() *EncodeStats {
	s := &EncodeStats{}
	e.Lock()
	if e.lastStats != nil {
		*s = *e.lastStats
	}
	e.Unlock()

	return s
}

// Options Devuelve las opciones que se esta usando para la codificacion
func (e *encodeSessionimpl) Options() *EncodeOptions {
	return e.options
}

// Cleanup limpia la sesión de codificación, descartando todos los fotogramas no leídos y deteniendo ffmpeg
// asegurando que ningún proceso ffmpeg comience a acumularse en su sistema
// acordate siempre que tenes que llamar a esto después de que esté hecho
func (e *encodeSessionimpl) Cleanup() {
	e.Stop()

	for _ = range e.frameChannel {

	}
}

// Read lee datos desde el búfer interno del EncodeSession en el slice proporcionado p.
// Implementa la interfaz io.Reader para permitir la lectura de frames de audio como si fuera un flujo de datos.
//
// Parámetros:
// - p: Un slice de bytes donde se almacenarán los datos leídos.
//
// Retorna:
// - El número de bytes leídos y almacenados en p.
// - Un error si ocurre algún problema durante la lectura, o nil si la operación es exitosa.
//
// Detalles del funcionamiento:
// - Si el búfer interno tiene suficientes datos para llenar el slice p, lee directamente desde el búfer.
// - Si el búfer no tiene suficientes datos, lee frames adicionales del canal de frames hasta que haya suficiente contenido en el búfer.
// - Cada frame leído del canal se escribe en el búfer interno para su posterior lectura.
// - Si se encuentra con un error al leer un frame, se registra el error y se retorna el error.
// - Cuando se alcanza el final del archivo (io.EOF), el método deja de leer más frames y continúa con los datos disponibles en el búfer.
// - Finalmente, lee los datos del búfer interno y los copia en el slice p.
func (e *encodeSessionimpl) Read(p []byte) (n int, err error) {
	if e.buf.Len() >= len(p) {
		return e.buf.Read(p)
	}

	for e.buf.Len() < len(p) {
		f, err := e.ReadFrame()
		if err != nil {
			break
		}
		e.buf.Write(f)
	}

	return e.buf.Read(p)
}

// FrameDuration implementa OpusReader, volviendo a ejecutar la duración de cada frame
func (e *encodeSessionimpl) FrameDuration() time.Duration {
	return time.Duration(e.options.FrameDuration) * time.Millisecond
}

// Error returns any error that occured during the encoding process
func (e *encodeSessionimpl) Error() error {
	e.Lock()
	defer e.Unlock()
	return e.err
}

// FFMPEGMessages devuelve los mensajes de salida de ffmpeg capturados durante la sesión de codificación.
//
// Retorna:
// - Un string que contiene los mensajes de salida de ffmpeg.
//
// Detalles del funcionamiento:
// - Adquiere un bloqueo en el mutex para garantizar un acceso seguro a la variable de salida de ffmpeg.
// - Copia el contenido de la variable de salida de ffmpeg a una variable local.
// - Libera el bloqueo en el mutex.
// - Devuelve el contenido de los mensajes de salida de ffmpeg.
func (e *encodeSessionimpl) FFMPEGMessages() string {
	e.Lock()
	output := e.ffmpegOutput
	e.Unlock()
	return output
}
