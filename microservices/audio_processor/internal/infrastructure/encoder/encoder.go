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
		Bitrate:          64,                    // Tasa de bits en kb/s (por ej. 64 kb/s)
		Application:      AudioApplicationAudio, // Aplicación de audio a usar
		CompressionLevel: 10,                    // Nivel de compresión (0 a 10, donde 10 es la máxima compresión y menor velocidad de codificación)
		PacketLoss:       1,                     // Porcentaje de pérdida de paquetes esperado
		BufferedFrames:   100,                   // Tamaño del búfer de cuadros
		VBR:              true,                  // Si se usa VBR (tasa de bits variable) o no
		StartTime:        0,                     // Tiempo de inicio de la secuencia de entrada en segundos
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
func EncodeMem(r io.Reader, options *EncodeOptions, ctx context.Context, logging logger.Logger) (session *EncodeSession, err error) {
	err = options.Validate()
	if err != nil {
		return
	}

	session = &EncodeSession{
		options:      options,
		pipeReader:   r,
		frameChannel: make(chan *Frame, options.BufferedFrames),
		logging:      logging,
	}
	go session.run(ctx)
	return
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
func (e *EncodeSession) run(ctx context.Context) {
	// Asegura que se marque la sesión como no en ejecución al finalizar.
	defer func() {
		e.Lock()
		e.running = false
		e.Unlock()
	}()

	// Marca la sesión como en ejecución y asegura el cierre del canal de frames.
	e.Lock()
	e.running = true
	defer close(e.frameChannel)
	e.Unlock()

	// Define el archivo de entrada. Usa "pipe:0" si filePath está vacío.
	inFile := "pipe:0"
	if e.filePath != "" {
		inFile = e.filePath
	}

	// Construye los argumentos para el comando ffmpeg.
	args := buildFFMPEGArgs(e.options, inFile)
	ffmpeg := exec.CommandContext(ctx, "ffmpeg", args...)
	e.logging.Debug("Ejecutando ffmpeg", zap.Strings("args", ffmpeg.Args), zap.String("input_file", inFile))

	// Configura el stdin de ffmpeg si se proporciona un pipeReader.
	if e.pipeReader != nil {
		ffmpeg.Stdin = e.pipeReader
	}

	// Obtiene pipes para stdout y stderr de ffmpeg.
	stdout, err := ffmpeg.StdoutPipe()
	if err != nil {
		e.logging.Error("Error al obtener stdout de ffmpeg", zap.Error(err))
		e.setError(ErrFailedToReadStdout)
		return
	}

	stderr, err := ffmpeg.StderrPipe()
	if err != nil {
		e.logging.Error("Error al obtener stderr de ffmpeg", zap.Error(err))
		e.setError(ErrFailedToReadStderr)
		return
	}

	// Escribe el marco de metadatos si no se utiliza salida en crudo.
	if !e.options.RawOutput {
		e.writeMetadataFrame()
	}

	// Inicia el proceso ffmpeg.
	err = ffmpeg.Start()
	if err != nil {
		e.logging.Error("Error al iniciar ffmpeg", zap.Error(err))
		e.setError(ErrFailedToStartFFMPEG)
		return
	}

	// Marca la hora de inicio y almacena el proceso.
	e.started = time.Now()
	e.process = ffmpeg.Process

	// Usa un grupo de espera para leer stderr en una goroutine.
	var wg sync.WaitGroup
	wg.Add(1)
	go e.readStderr(stderr, &wg)

	// Lee stdout del proceso.
	e.readStdout(stdout)
	wg.Wait()

	// Espera a que ffmpeg termine y maneja cualquier error.
	if err := ffmpeg.Wait(); err != nil && err.Error() != "signal: killed" {
		e.logging.Error("Error al esperar a ffmpeg", zap.Error(err))
		e.setError(err)
	}
}

// buildFFMPEGArgs construye una lista de argumentos para el comando ffmpeg basado en las opciones de codificación proporcionadas.
//
// Parámetros:
// - options: Opciones de codificación que determinan cómo se debe procesar el audio.
// - inFile: Ruta del archivo de entrada o "pipe:0" si se utiliza un pipe para la entrada.
//
// Retorna:
// - Una lista de argumentos para el comando ffmpeg.
//
// Detalles del funcionamiento:
// - Configura parámetros para la reconexión automática, códec de audio, formato de salida, tasa de bits, volumen, frecuencia de muestreo, etc.
// - Convierte las opciones de configuración en una cadena de argumentos que ffmpeg puede interpretar.
func buildFFMPEGArgs(options *EncodeOptions, inFile string) []string {
	return []string{
		"-stats",     // Muestra estadísticas durante el procesamiento
		"-i", inFile, // Archivo de entrada
		"-reconnect", "1", // Habilita la reconexión automática
		"-reconnect_at_eof", "1", // Reconecta al final del archivo
		"-reconnect_streamed", "1", // Reconecta para transmisiones en vivo
		"-reconnect_delay_max", "2", // Tiempo máximo de retraso para la reconexión
		"-map", "0:a", // Mapea la primera pista de audio
		"-acodec", "libopus", // Utiliza el códec de audio libopus
		"-f", "ogg", // Establece el formato de salida a OGG
		"-vbr", boolToStr(options.VBR), // Establece si se usa VBR (tasa de bits variable)
		"-compression_level", strconv.Itoa(options.CompressionLevel), // Nivel de compresión
		"-af", fmt.Sprintf("volume=%.2f", float64(options.Volume)/100.0), // Ajusta el volumen
		"-ar", strconv.Itoa(options.FrameRate), // Frecuencia de muestreo del audio en Hz
		"-ac", strconv.Itoa(options.Channels), // Número de canales de audio
		"-b:a", strconv.Itoa(options.Bitrate * 1000), // Tasa de bits de audio en bps
		"-application", string(options.Application), // Aplicación de audio (por ej. "audio")
		"-frame_duration", strconv.Itoa(options.FrameDuration), // Duración del marco en ms
		"-packet_loss", strconv.Itoa(options.PacketLoss), // Porcentaje de pérdida de paquetes
		"-threads", strconv.Itoa(options.Threads), // Número de hilos a utilizar
		"-ss", strconv.Itoa(options.StartTime), // Tiempo de inicio en segundos
		"pipe:1", // Salida a través de un pipe
	}
}

func boolToStr(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

// setError asigna un error a la sesión de codificación si no hay un error previamente registrado.
//
// Parámetros:
// - err: El error que se debe asignar a la sesión.
//
// Detalles del funcionamiento:
// - Utiliza un mutex para garantizar que la asignación del error sea segura en entornos concurrentes.
// - Solo asigna el error si no se ha registrado ningún otro error previamente.
//
// Este método asegura que solo se registre un error, evitando sobrescribir errores anteriores.
func (e *EncodeSession) setError(err error) {
	e.Lock()          // Adquiere el mutex para acceso exclusivo.
	defer e.Unlock()  // Libera el mutex al finalizar.
	if e.err == nil { // Verifica si ya hay un error registrado.
		e.err = err // Asigna el nuevo error solo si no hay error registrado.
	}
}

// readStderr lee la salida de error (stderr) del proceso ffmpeg y maneja cada línea de error.
// Este método se ejecuta en una goroutine para permitir la lectura concurrente de stderr.
//
// Parámetros:
// - stderr: Un io.ReadCloser que proporciona la salida de error del proceso ffmpeg.
// - wg: Un WaitGroup para sincronizar la finalización de la lectura de stderr.
//
// Detalles del funcionamiento:
// - Utiliza un búfer para leer los datos de stderr carácter por carácter.
// - Acumula caracteres en un búfer hasta encontrar un carácter de nueva línea o retorno de carro.
// - Cuando se encuentra un retorno de carro ('\r'), maneja la línea acumulada y la limpia.
// - Cuando se encuentra una nueva línea ('\n'), agrega la línea al búfer de salida de ffmpeg y la limpia.
// - Registra errores si ocurren durante la lectura, excepto cuando se alcanza el final del archivo (EOF).
func (e *EncodeSession) readStderr(stderr io.ReadCloser, wg *sync.WaitGroup) {
	defer wg.Done() // Señala que la goroutine ha terminado cuando se sale de la función.

	bufReader := bufio.NewReader(stderr) // Crea un lector de búfer para leer de stderr.
	var outBuf bytes.Buffer              // Búfer para acumular caracteres de la salida de error.

	for {
		// Lee un carácter de stderr.
		r, _, err := bufReader.ReadRune()
		if err != nil {
			if err != io.EOF { // Registra un error si no es EOF.
				e.logging.Error("Error leyendo stderr", zap.Error(err))
			}
			break // Sale del bucle en caso de error o EOF.
		}

		switch r {
		case '\r': // Si el carácter es un retorno de carro.
			if outBuf.Len() > 0 { // Si hay datos en el búfer.
				e.handleStderrLine(outBuf.String()) // Maneja la línea completa.
				outBuf.Reset()                      // Limpia el búfer.
			}
		case '\n': // Si el carácter es una nueva línea.
			e.Lock()                                 // Adquiere el mutex para acceso seguro.
			e.ffmpegOutput += outBuf.String() + "\n" // Agrega la línea al búfer de salida de ffmpeg.
			e.Unlock()                               // Libera el mutex.
			outBuf.Reset()                           // Limpia el búfer.
		default:
			outBuf.WriteRune(r) // Acumula caracteres en el búfer.
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
func (e *EncodeSession) handleStderrLine(line string) {
	if strings.Index(line, "size=") != 0 { // Verifica si la línea contiene información relevante.
		return
	}

	var size int
	var timeH int
	var timeM int
	var timeS float32
	var bitrate float32
	var speed float32

	// Analiza la línea y extrae el tamaño, tiempo, tasa de bits y velocidad.
	_, err := fmt.Sscanf(line, "size=%dkB time=%d:%d:%f bitrate=%fkbits/s speed=%fx", &size, &timeH, &timeM, &timeS, &bitrate, &speed)
	if err != nil {
		e.logging.Error("Error al analizar línea de stderr", zap.Error(err)) // Registra un error si el análisis falla.
	}

	// Calcula la duración total en base a las horas, minutos y segundos extraídos.
	dur := time.Duration(timeH) * time.Hour
	dur += time.Duration(timeM) * time.Minute
	dur += time.Duration(timeS) * time.Second

	// Crea una instancia de EncodeStats con la información extraída.
	stats := &EncodeStats{
		Size:     size,
		Duration: dur,
		Bitrate:  bitrate,
		Speed:    speed,
	}

	// Utiliza un mutex para actualizar lastStats de manera segura.
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
func (e *EncodeSession) readStdout(stdout io.ReadCloser) {
	decoder := NewPacketDecoder(ogg.NewDecoder(stdout)) // Crea un decodificador para los paquetes OGG desde stdout.

	skipPackets := 2 // Número de paquetes a omitir al inicio.
	for {
		// Decodifica un paquete desde stdout.
		packet, _, err := decoder.Decode()
		if skipPackets > 0 {
			skipPackets-- // Omite los primeros dos paquetes.
			continue
		}
		if err != nil {
			if err != io.EOF { // Registra un error si no es EOF.
				e.logging.Error("Error al leer stdout", zap.Error(err))
			}
			break // Sale del bucle en caso de error o EOF.
		}
		// Escribe el paquete decodificado en formato Opus.
		err = e.writeOpusFrame(packet)
		if err != nil {
			e.logging.Error("Error escribir opus frame", zap.Error(err)) // Registra un error si ocurre durante la escritura.
			break                                                        // Sale del bucle en caso de error.
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
func (e *EncodeSession) writeOpusFrame(opusFrame []byte) error {
	var dcaBuf bytes.Buffer // Búfer para almacenar los datos en formato DCA.

	// Escribe el tamaño del frame Opus como un entero de 16 bits en el búfer.
	err := binary.Write(&dcaBuf, binary.LittleEndian, int16(len(opusFrame)))
	if err != nil {
		e.logging.Error("Error al escribir datos de frame DCA", zap.Error(err)) // Registra un error si ocurre durante la escritura.
		return err
	}

	// Escribe el frame Opus en el búfer.
	_, err = dcaBuf.Write(opusFrame)
	if err != nil {
		e.logging.Error("Error al escribir frame Opus", zap.Error(err)) // Registra un error si ocurre durante la escritura.
		return err
	}

	// Envía el búfer con los datos del frame a través del canal de frames.
	e.frameChannel <- &Frame{dcaBuf.Bytes(), false}

	e.Lock()      // Adquiere el mutex para acceso seguro.
	e.lastFrame++ // Incrementa el contador de frames.
	e.Unlock()    // Libera el mutex.

	return nil // Retorna nil si la operación es exitosa.
}

// writeMetadataFrame escribe un frame de metadatos en el canal de frames.
// Este método crea un frame que contiene información sobre la configuración de la codificación y la fuente del archivo.
//
// Detalles del funcionamiento:
// - Crea una estructura de metadatos que incluye información sobre la codificación Opus y el origen del archivo.
// - La información de metadatos incluye la tasa de bits, la frecuencia de muestreo, la aplicación utilizada, el tamaño del frame, el número de canales y si se utiliza VBR.
// - Serializa los metadatos a formato JSON.
// - Escribe un encabezado "DCA1" seguido de la longitud del JSON en formato Little Endian y luego el JSON serializado en un búfer.
// - Envía el búfer que contiene el frame de metadatos al canal de frames.
// - Registra errores si ocurren durante la codificación de JSON o la escritura de datos y retorna sin enviar el frame en caso de error.
func (e *EncodeSession) writeMetadataFrame() {
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
		e.logging.Error("Error al codificar metadatos en JSON", zap.Error(err)) // Registra un error si la serialización falla.
		return
	}

	// Escribe la longitud del JSON en formato Little Endian.
	if err := binary.Write(&buf, binary.LittleEndian, int32(len(jsonData))); err != nil {
		e.logging.Error("Error al escribir longitud de JSON", zap.Error(err)) // Registra un error si la escritura falla.
		return
	}

	// Escribe el JSON serializado en el búfer.
	buf.Write(jsonData)
	// Envía el búfer con el frame de metadatos al canal de frames.
	e.frameChannel <- &Frame{buf.Bytes(), true}
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
func (e *EncodeSession) ReadFrame() (frame []byte, err error) {
	f := <-e.frameChannel // Lee un frame del canal de frames.
	if f == nil {
		return nil, io.EOF // Retorna io.EOF si el canal está cerrado y no hay más frames.
	}

	return f.data, nil // Retorna los datos del frame.
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
func (e *EncodeSession) Read(p []byte) (n int, err error) {
	if e.buf.Len() >= len(p) {
		// Si el búfer tiene suficientes datos, lee directamente desde el búfer.
		return e.buf.Read(p)
	}

	for e.buf.Len() < len(p) {
		// Si el búfer no tiene suficientes datos, lee frames adicionales.
		f, err := e.ReadFrame()
		if err != nil {
			if err == io.EOF {
				// Si se llega al final del archivo, sale del bucle.
				break
			}
			e.logging.Error("Error al leer frame", zap.Error(err)) // Registra un error si ocurre.
			return 0, err
		}
		e.buf.Write(f) // Escribe el frame en el búfer interno.
	}
	return e.buf.Read(p) // Lee los datos del búfer y los almacena en p.
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
func (e *EncodeSession) FFMPEGMessages() string {
	e.Lock()                 // Adquiere el mutex para acceso seguro a la variable de salida de ffmpeg.
	output := e.ffmpegOutput // Copia el contenido de los mensajes de salida de ffmpeg.
	e.Unlock()               // Libera el mutex.
	return output            // Devuelve el contenido de los mensajes de salida.
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
func (e *EncodeSession) Stop() error {
	e.Lock()         // Adquiere el mutex para acceso seguro a los atributos de la sesión.
	defer e.Unlock() // Libera el mutex al finalizar la función.

	if !e.running || e.process == nil {
		return errors.New("la session no esta corriendo") // Retorna un error si la sesión no está en ejecución.
	}
	if err := e.process.Kill(); err != nil {
		e.logging.Error("Error al detener el proceso de codificación", zap.Error(err)) // Registra un error si no se puede detener el proceso.
		return err
	}
	e.running = false // Actualiza el estado de la sesión para indicar que no está en ejecución.
	return nil        // Retorna nil si la operación es exitosa.
}
