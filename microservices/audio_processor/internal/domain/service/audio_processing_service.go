package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/port"
	"github.com/cenkalti/backoff/v4"
	"io"
	"time"

	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/api"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/downloader"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/encoder"
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
	audioFileExtension = ".dca"      // Extensión de archivo para los audios procesados.
	maxBackoff         = 30 * time.Second
)

// AudioProcessingService es un servicio que maneja la descarga, codificación y almacenamiento de audio.
type (
	AudioProcessingService struct {
		log            logger.Logger            // Logger para el registro de eventos y errores.
		storage        port.Storage             // Interfaz para el almacenamiento en S3.
		downloader     downloader.Downloader    // Interfaz para la descarga de audio.
		operationStore port.OperationRepository // Interfaz para almacenar resultados de operaciones.
		encoder        encoder.AudioEncoder
		metadataStore  port.MetadataRepository // Interfaz para almacenar metadatos del audio.
		messaging      port.MessageQueue       // Interfaz para enviar mensajes a un message broker
		config         *config.Config          // Configuración del servicio.
	}

	AudioProcessor interface {
		StartOperation(ctx context.Context, song string) (string, string, error)
		ProcessAudio(ctx context.Context, operationID string, metadata *api.VideoDetails) error
	}
)

// NewAudioProcessingService crea una nueva instancia de AudioProcessingService con las configuraciones proporcionadas.
func NewAudioProcessingService(log logger.Logger, storage port.Storage,
	downloader downloader.Downloader,
	operationStore port.OperationRepository,
	metadataStore port.MetadataRepository,
	messaging port.MessageQueue,
	encoder encoder.AudioEncoder,
	config *config.Config) *AudioProcessingService {

	if config.Service.MaxAttempts == 0 {
		config.Service.MaxAttempts = maxAttempts
	}
	if config.Service.Timeout == 0 {
		config.Service.Timeout = 5 * time.Minute
	}

	return &AudioProcessingService{
		log:            log,
		storage:        storage,
		downloader:     downloader,
		operationStore: operationStore,
		metadataStore:  metadataStore,
		messaging:      messaging,
		encoder:        encoder,
		config:         config,
	}
}

// StartOperation inicia una nueva operación de procesamiento de audio y guarda su estado inicial.
func (a *AudioProcessingService) StartOperation(ctx context.Context, songID string) (string, string, error) {
	operationResult := &model.OperationResult{
		ID:     uuid.New().String(),
		SK:     songID,
		Status: statusInitiating,
	}

	if err := a.operationStore.SaveOperationsResult(ctx, operationResult); err != nil {
		return "", "", fmt.Errorf("error al guardar resultado de operacion: %w", err)
	}
	return operationResult.ID, operationResult.SK, nil
}

// ProcessAudio procesa el audio descargando, codificando y almacenando en S3, con reintentos en caso de fallos.
func (a *AudioProcessingService) ProcessAudio(ctx context.Context, operationID string, youtubeMetadata *api.VideoDetails) error {
	ctx, cancel := context.WithTimeout(ctx, a.config.Service.Timeout)
	defer cancel()

	metadata := a.createMetadata(youtubeMetadata)

	operation := func() error {
		return a.processAudioAttempt(ctx, operationID, metadata)
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = a.config.Service.Timeout
	bo.MaxInterval = maxBackoff

	if err := backoff.Retry(operation, bo); err != nil {
		a.log.Error("Fallo despues de varios intentos", zap.Error(err))
		return a.handleFailedProcessing(ctx, operationID, metadata)
	}
	return nil
}

// processAudioAttempt intenta procesar el audio una vez, manejando la descarga, codificación y almacenamiento.
func (a *AudioProcessingService) processAudioAttempt(ctx context.Context, operationID string, metadata *model.Metadata) error {
	reader, err := a.downloader.DownloadAudio(ctx, metadata.URLYouTube)
	if err != nil {
		a.log.Error("Error al descargar audio",
			zap.String("operationID", operationID),
			zap.Error(err))
		return fmt.Errorf("error al descargar audio: %w", err)
	}
	defer reader.Close()

	session, err := a.encoder.Encode(ctx, reader, encoder.StdEncodeOptions)
	if err != nil {
		a.log.Error("Error al codificar audio",
			zap.String("operationID", operationID),
			zap.Error(err))
		return fmt.Errorf("error al codificar audio: %w", err)
	}
	defer session.Stop()

	frames, err := a.readAudioFramesToBuffer(session)
	if err != nil {
		a.log.Error("Error al leer frames de audio",
			zap.String("operationID", operationID),
			zap.Error(err))
		return fmt.Errorf("error al leer los frames: %w", err)
	}

	if frames.Len() == 0 {
		return fmt.Errorf("buffer de audio vacio")
	}

	keyName := fmt.Sprintf("%s%s", metadata.Title, audioFileExtension)
	err = a.storage.UploadFile(ctx, keyName, frames)
	if err != nil {
		a.log.Error("Error al subir archivo",
			zap.String("operationID", operationID),
			zap.Error(err))
		return fmt.Errorf("error en guardar el archivo: %w", err)
	}

	fileMetadata, err := a.storage.GetFileMetadata(ctx, keyName)
	if err != nil {
		return fmt.Errorf("error al obtener metadata del archivo: %w", err)
	}

	err = a.metadataStore.SaveMetadata(ctx, metadata)
	if err != nil {
		return fmt.Errorf("error al guardar metadata: %w", err)
	}

	result := a.createSuccessResult(operationID, metadata, fileMetadata)
	err = a.operationStore.UpdateOperationResult(ctx, operationID, result)
	if err != nil {
		return fmt.Errorf("error al guardar resultado de operación: %w", err)
	}

	message := model.Message{
		ID:      operationID,
		Content: "Procesamiento de audio exitoso",
		Status: model.Status{
			ID:             operationID,
			SK:             metadata.VideoID,
			Status:         statusSuccess,
			Message:        "Procesamiento exitoso",
			Metadata:       metadata,
			FileData:       fileMetadata,
			ProcessingDate: time.Now().UTC(),
			Success:        true,
			Attempts:       1,
			Failures:       0,
		},
	}

	if err := a.messaging.SendMessage(ctx, message); err != nil {
		a.log.Error("Error al enviar el mensaje de exito", zap.Error(err))
		return err
	}

	a.log.Info("Procesamiento exitoso")
	return nil
}

// createMetadata genera metadatos a partir de los detalles de un video de YouTube.
func (a *AudioProcessingService) createMetadata(youtubeMetadata *api.VideoDetails) *model.Metadata {
	return &model.Metadata{
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
func (a *AudioProcessingService) createSuccessResult(operationID string, metadata *model.Metadata, fileData *model.FileData) *model.OperationResult {
	return &model.OperationResult{
		ID:             operationID,
		SK:             metadata.VideoID,
		Status:         statusSuccess,
		Message:        "Procesamiento exitoso",
		Metadata:       metadata,
		FileData:       fileData,
		ProcessingDate: time.Now().Format(time.RFC3339),
		Success:        true,
		Attempts:       1,
		Failures:       0,
	}
}

// handleFailedProcessing maneja el caso en que el procesamiento falla después de varios intentos.
func (a *AudioProcessingService) handleFailedProcessing(ctx context.Context, operationID string, metadata *model.Metadata) error {
	result := &model.OperationResult{
		ID:             operationID,
		SK:             metadata.VideoID,
		Status:         statusFailed,
		Metadata:       metadata,
		Message:        fmt.Sprintf("Fallo en el procesamiento después de varios intentos: %d", a.config.Service.MaxAttempts),
		ProcessingDate: time.Now().Format(time.RFC3339),
		Success:        false,
		Attempts:       maxAttempts,
		Failures:       maxAttempts,
	}

	if err := a.operationStore.UpdateOperationResult(ctx, operationID, result); err != nil {
		return fmt.Errorf("error al guardar resultado de operación fallida: %w", err)
	}

	message := model.Message{
		ID:      operationID,
		Content: "Procesamiento fallido",
		Status: model.Status{
			ID:             operationID,
			SK:             metadata.VideoID,
			Status:         statusFailed,
			Message:        "El procesamiento falló después de varios intentos",
			Metadata:       metadata,
			ProcessingDate: time.Now().UTC(),
			Success:        false,
			Attempts:       a.config.Service.MaxAttempts,
			Failures:       a.config.Service.MaxAttempts,
		},
	}

	if err := a.messaging.SendMessage(ctx, message); err != nil {
		a.log.Error("Error al enviar el mensaje de fallo", zap.Error(err))
		return err
	}

	a.log.Error("Procesamiento fallido después de varios intentos")
	return fmt.Errorf("procesamiento fallido después de varios intentos")
}

// readAudioFramesToBuffer lee los frames de audio de la sesión de codificación y los almacena en un buffer.
func (a *AudioProcessingService) readAudioFramesToBuffer(session encoder.EncodeSession) (*bytes.Buffer, error) {
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
