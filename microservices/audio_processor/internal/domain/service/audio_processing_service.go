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
	maxAttempts        = 1           // Número máximo de intentos permitidos para procesar audio.
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
	var reader io.Reader
	attempts := 0

	// Función que ejecuta todo el proceso y maneja los errores
	operation := func() error {
		attempts++
		var err error

		reader, err = a.downloader.DownloadAudio(ctx, youtubeMetadata.URLYouTube)
		if err != nil {
			a.log.Error("Error en la descarga de audio", zap.Error(err))
			// Guardar el estado del error de descarga
			if saveErr := a.saveErrorState(ctx, operationID, metadata, "Error en descarga", attempts, err); saveErr != nil {
				a.log.Error("Error al guardar estado de error de descarga", zap.Error(saveErr))
			}
			return err
		}

		session, err := a.encoder.Encode(ctx, reader, encoder.StdEncodeOptions)
		if err != nil {
			if saveErr := a.saveErrorState(ctx, operationID, metadata, "Error en codificación", attempts, err); saveErr != nil {
				a.log.Error("Error al guardar estado de error de codificación", zap.Error(saveErr))
			}
			return err
		}
		if session != nil {
			defer session.Cleanup()
		}

		frames, err := a.readAudioFramesToBuffer(session)
		if err != nil {
			if saveErr := a.saveErrorState(ctx, operationID, metadata, "Error en lectura de frames", attempts, err); saveErr != nil {
				a.log.Error("Error al guardar estado de error de lectura", zap.Error(saveErr))
			}
			return err
		}

		if frames.Len() == 0 {
			err = fmt.Errorf("buffer de audio vacío")
			if saveErr := a.saveErrorState(ctx, operationID, metadata, "Buffer vacío", attempts, err); saveErr != nil {
				a.log.Error("Error al guardar estado de buffer vacío", zap.Error(saveErr))
			}
			return err
		}

		keyName := fmt.Sprintf("%s%s", metadata.Title, audioFileExtension)
		err = a.storage.UploadFile(ctx, keyName, frames)
		if err != nil {
			if saveErr := a.saveErrorState(ctx, operationID, metadata, "Error en subida a S3", attempts, err); saveErr != nil {
				a.log.Error("Error al guardar estado de error de subida", zap.Error(saveErr))
			}
			return err
		}

		fileMetadata, err := a.storage.GetFileMetadata(ctx, keyName)
		if err != nil {
			if saveErr := a.saveErrorState(ctx, operationID, metadata, "Error al obtener metadata", attempts, err); saveErr != nil {
				a.log.Error("Error al guardar estado de error de metadata", zap.Error(saveErr))
			}
			return err
		}

		if err = a.metadataStore.SaveMetadata(ctx, metadata); err != nil {
			if saveErr := a.saveErrorState(ctx, operationID, metadata, "Error al guardar metadata", attempts, err); saveErr != nil {
				a.log.Error("Error al guardar estado de error de guardado de metadata", zap.Error(saveErr))
			}
			return err
		}

		result := a.createSuccessResult(operationID, metadata, fileMetadata)
		if err = a.operationStore.UpdateOperationResult(ctx, operationID, result); err != nil {
			if saveErr := a.saveErrorState(ctx, operationID, metadata, "Error al actualizar resultado", attempts, err); saveErr != nil {
				a.log.Error("Error al guardar estado de error de actualización", zap.Error(saveErr))
			}
			return err
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
				Attempts:       attempts,
				Failures:       attempts - 1,
			},
		}

		if err = a.messaging.SendMessage(ctx, message); err != nil {
			if saveErr := a.saveErrorState(ctx, operationID, metadata, "Error al enviar mensaje", attempts, err); saveErr != nil {
				a.log.Error("Error al guardar estado de error de envío", zap.Error(saveErr))
			}
			return err
		}

		return nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = a.config.Service.Timeout
	bo.MaxInterval = maxBackoff

	if err := backoff.Retry(operation, bo); err != nil {
		a.log.Error("Fallo después de varios intentos", zap.Error(err))
		return a.handleFailedProcessing(ctx, operationID, metadata)
	}

	return nil
}

func (a *AudioProcessingService) processAudioAttemptDownload(ctx context.Context, operationID string, metadata *model.Metadata) (io.Reader, error) {
	reader, err := a.downloader.DownloadAudio(ctx, metadata.URLYouTube)
	if err != nil {
		a.log.Error("Error al descargar audio",
			zap.String("operationID", operationID),
			zap.Error(err))
		return nil, fmt.Errorf("error al descargar audio: %w", err)
	}
	return reader, nil
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

func (a *AudioProcessingService) saveErrorState(ctx context.Context, operationID string, metadata *model.Metadata, errorMsg string, attempts int, originalErr error) error {
	result := &model.OperationResult{
		ID:             operationID,
		SK:             metadata.VideoID,
		Status:         "error",
		Message:        fmt.Sprintf("%s: %v", errorMsg, originalErr),
		Metadata:       metadata,
		ProcessingDate: time.Now().Format(time.RFC3339),
		Success:        false,
		Attempts:       attempts,
		Failures:       attempts,
	}

	// Guardar el resultado de la operación
	if err := a.operationStore.UpdateOperationResult(ctx, operationID, result); err != nil {
		return err
	}

	// Enviar mensaje de error
	message := model.Message{
		ID:      operationID,
		Content: "Error en procesamiento",
		Status: model.Status{
			ID:             operationID,
			SK:             metadata.VideoID,
			Status:         "error",
			Message:        fmt.Sprintf("%s: %v", errorMsg, originalErr),
			Metadata:       metadata,
			ProcessingDate: time.Now().UTC(),
			Success:        false,
			Attempts:       attempts,
			Failures:       attempts,
		},
	}

	return a.messaging.SendMessage(ctx, message)
}
