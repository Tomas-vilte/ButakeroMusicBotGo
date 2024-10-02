package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"io"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/repository"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/downloader"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/encoder"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/storage"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	maxAttempts        = 3           // Número máximo de intentos permitidos para procesar audio.
	statusInitiating   = "iniciando" // Estado inicial de la operación.
	statusSuccess      = "success"   // Estado de la operación cuando se procesa con éxito.
	statusFailed       = "failed"    // Estado de la operación cuando falla después de intentos.
	platformYoutube    = "Youtube"   // Plataforma de origen del audio.
	s3BucketName       = "butakero"  // Nombre del bucket en S3 donde se almacenarán los archivos.
	audioFileExtension = ".dca"      // Extensión de archivo para los audios procesados.
)

// AudioProcessingService es un servicio que maneja la descarga, codificación y almacenamiento de audio.
type (
	AudioProcessingService struct {
		log            logger.Logger                  // Logger para el registro de eventos y errores.
		storage        storage.Storage                // Interfaz para el almacenamiento en S3.
		downloader     downloader.Downloader          // Interfaz para la descarga de audio.
		operationStore repository.OperationRepository // Interfaz para almacenar resultados de operaciones.
		metadataStore  repository.MetadataRepository  // Interfaz para almacenar metadatos del audio.
		config         config.Config                  // Configuración del servicio.
	}

	AudioProcessor interface {
		StartOperation(ctx context.Context, song string) (string, string, error)
		ProcessAudio(ctx context.Context, operationID string, metadata api.VideoDetails) error
	}
)

// NewAudioProcessingService crea una nueva instancia de AudioProcessingService con las configuraciones proporcionadas.
func NewAudioProcessingService(log logger.Logger, storage storage.Storage,
	downloader downloader.Downloader,
	operationStore repository.OperationRepository,
	metadataStore repository.MetadataRepository,
	config config.Config) *AudioProcessingService {

	if config.MaxAttempts == 0 {
		config.MaxAttempts = maxAttempts
	}
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Minute
	}

	return &AudioProcessingService{
		log:            log,
		storage:        storage,
		downloader:     downloader,
		operationStore: operationStore,
		metadataStore:  metadataStore,
		config:         config,
	}
}

// StartOperation inicia una nueva operación de procesamiento de audio y guarda su estado inicial.
func (a *AudioProcessingService) StartOperation(ctx context.Context, songID string) (string, string, error) {
	operationResult := model.OperationResult{
		ID:     uuid.New().String(),
		SongID: songID,
		Status: statusInitiating,
	}

	err := a.operationStore.SaveOperationsResult(ctx, operationResult)
	if err != nil {
		return "", "", fmt.Errorf("error al guardar operación: %w", err)
	}
	return operationResult.ID, operationResult.SongID, nil
}

// ProcessAudio procesa el audio descargando, codificando y almacenando en S3, con reintentos en caso de fallos.
func (a *AudioProcessingService) ProcessAudio(ctx context.Context, operationID string, youtubeMetadata api.VideoDetails) error {
	ctx, cancel := context.WithTimeout(ctx, a.config.Timeout)
	defer cancel()

	metadata := a.createMetadata(youtubeMetadata)

	for attempts := 1; attempts <= a.config.MaxAttempts; attempts++ {
		err := a.processAudioAttempt(ctx, operationID, metadata, attempts)
		if err == nil {
			return nil
		}
		a.log.Error("Attempt failed", zap.Int("attempt", attempts), zap.Error(err))
	}

	return a.handleFailedProcessing(ctx, operationID, metadata)
}

// processAudioAttempt intenta procesar el audio una vez, manejando la descarga, codificación y almacenamiento.
func (a *AudioProcessingService) processAudioAttempt(ctx context.Context, operationID string, metadata model.Metadata, attempt int) error {
	reader, err := a.downloader.DownloadAudio(ctx, metadata.URLYouTube)
	if err != nil {
		return fmt.Errorf("error al descargar audio: %w", err)
	}

	session, err := encoder.EncodeMem(reader, encoder.StdEncodeOptions, ctx, a.log)
	if err != nil {
		return fmt.Errorf("error al codificar audio: %w", err)
	}

	frames, err := a.readAudioFramesToBuffer(session)
	if err != nil {
		return fmt.Errorf("error al leer los frames: %w", err)
	}

	keyS3 := fmt.Sprintf("%s%s", metadata.Title, audioFileExtension)
	err = a.storage.UploadFile(ctx, keyS3, frames)
	if err != nil {
		return fmt.Errorf("error al subir archivo a S3: %w", err)
	}

	metadata.URLS3 = fmt.Sprintf("s3://%s/%s", s3BucketName, keyS3)

	err = a.metadataStore.SaveMetadata(ctx, metadata)
	if err != nil {
		return fmt.Errorf("error al guardar metadata: %w", err)
	}

	result := a.createSuccessResult(operationID, metadata, attempt)
	err = a.operationStore.SaveOperationsResult(ctx, result)
	if err != nil {
		return fmt.Errorf("error al guardar resultado de operación: %w", err)
	}

	a.log.Info("Procesamiento exitoso", zap.Int("attempts", attempt))
	return nil
}

// createMetadata genera metadatos a partir de los detalles de un video de YouTube.
func (a *AudioProcessingService) createMetadata(youtubeMetadata api.VideoDetails) model.Metadata {
	return model.Metadata{
		ID:         uuid.New().String(),
		VideoID:    youtubeMetadata.VideoID,
		Title:      youtubeMetadata.Title,
		Duration:   youtubeMetadata.Duration,
		URLYouTube: youtubeMetadata.URLYouTube,
		Platform:   platformYoutube,
		Thumbnail:  youtubeMetadata.Thumbnail,
	}
}

// createSuccessResult crea un resultado de operación exitoso después del procesamiento de audio.
func (a *AudioProcessingService) createSuccessResult(operationID string, metadata model.Metadata, attempts int) model.OperationResult {
	return model.OperationResult{
		ID:             operationID,
		SongID:         metadata.VideoID,
		Status:         statusSuccess,
		Message:        "Procesamiento exitoso",
		Data:           fmt.Sprintf("Archivo guardado en S3: %s", metadata.URLS3),
		ProcessingDate: time.Now().Format(time.RFC3339),
		Success:        true,
		Attempts:       attempts,
		Failures:       attempts - 1,
	}
}

// handleFailedProcessing maneja el caso en que el procesamiento falla después de varios intentos.
func (a *AudioProcessingService) handleFailedProcessing(ctx context.Context, operationID string, metadata model.Metadata) error {
	result := model.OperationResult{
		ID:             operationID,
		SongID:         metadata.VideoID,
		Status:         statusFailed,
		Message:        "Fallo en el procesamiento después de varios intentos",
		Data:           fmt.Sprintf("Fallos: %d", a.config.MaxAttempts),
		ProcessingDate: time.Now().Format(time.RFC3339),
		Success:        false,
		Attempts:       a.config.MaxAttempts,
		Failures:       a.config.MaxAttempts,
	}

	err := a.operationStore.SaveOperationsResult(ctx, result)
	if err != nil {
		a.log.Error("Error al guardar resultado de operación fallida", zap.Error(err))
	}

	a.log.Error("El procesamiento falló después de varios intentos", zap.Int("attempts", a.config.MaxAttempts))
	return fmt.Errorf("el procesamiento falló después de %d intentos", a.config.MaxAttempts)
}

// readAudioFramesToBuffer lee los frames de audio de la sesión de codificación y los almacena en un buffer.
func (a *AudioProcessingService) readAudioFramesToBuffer(session *encoder.EncodeSession) (*bytes.Buffer, error) {
	var buffer bytes.Buffer

	for {
		frame, err := session.ReadFrame()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error al leer frame de audio: %w", err)
		}

		_, err = buffer.Write(frame)
		if err != nil {
			return nil, fmt.Errorf("error al escribir frame en buffer: %w", err)
		}
	}
	return &buffer, nil
}
