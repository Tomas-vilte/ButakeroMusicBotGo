package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/cenkalti/backoff/v4"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"time"
)

const (
	statusInitiating   = "starting" // Estado inicial de la operación.
	statusSuccess      = "success"  // Estado de la operación cuando se procesa con éxito.
	statusFailed       = "failed"   // Estado de la operación cuando falla después de intentos.
	platformYoutube    = "Youtube"  // Plataforma de origen del audio.
	audioFileExtension = ".dca"     // Extensión de archivo para los audios procesados.
	maxBackoff         = 30 * time.Second
)

type AudioProcessingService struct {
	downloadService  ports.AudioDownloadService
	storageService   ports.AudioStorageService
	opsManager       ports.OperationsManager
	messagingService ports.MessagingManager
	errorHandler     ports.ErrorManagement
	logger           logger.Logger
	config           *config.Config
}

func NewAudioProcessingService(
	downloadService ports.AudioDownloadService,
	storageService ports.AudioStorageService,
	opsManager ports.OperationsManager,
	messagingService ports.MessagingManager,
	errorHandler ports.ErrorManagement,
	logger logger.Logger,
	cfg *config.Config,
) *AudioProcessingService {
	return &AudioProcessingService{
		downloadService:  downloadService,
		storageService:   storageService,
		opsManager:       opsManager,
		messagingService: messagingService,
		errorHandler:     errorHandler,
		logger:           logger,
		config:           cfg,
	}
}

// ProcessAudio procesa el audio descargando, codificando y almacenando en S3, con reintentos en caso de fallos.
func (a *AudioProcessingService) ProcessAudio(ctx context.Context, operationID string, mediaDetails *model.MediaDetails) error {
	ctx, cancel := context.WithTimeout(ctx, a.config.Service.Timeout)
	defer cancel()

	metadata := a.createMetadata(mediaDetails)
	attempts := 0

	a.logger.Info("Iniciando procesamiento de audio",
		zap.String("operation_id", operationID),
		zap.String("title", metadata.Title),
		zap.String("video_id", metadata.VideoID))

	operation := func() error {
		attempts++

		if attempts > a.config.Service.MaxAttempts {
			return backoff.Permanent(fmt.Errorf("número máximo de intentos alcanzado (%d)", a.config.Service.MaxAttempts))
		}

		if attempts > 1 {
			a.logger.Info("Reintentando procesamiento",
				zap.String("operation_id", operationID),
				zap.Int("attempt", attempts),
				zap.Int("max_attempts", a.config.Service.MaxAttempts))
		}

		audioBuffer, err := a.downloadService.DownloadAndEncode(ctx, mediaDetails.URL)
		if err != nil {
			a.logger.Error("Error en descarga/codificación",
				zap.String("operation_id", operationID),
				zap.Error(err))
			return a.errorHandler.HandleProcessingError(ctx, operationID, metadata, "download/encode", attempts, err)
		}

		a.logger.Debug("Audio descargado y codificado correctamente",
			zap.String("operation_id", operationID))

		fileData, err := a.storageService.StoreAudio(ctx, audioBuffer, metadata)
		if err != nil {
			a.logger.Error("Error al almacenar audio",
				zap.String("operation_id", operationID),
				zap.Error(err))
			return a.errorHandler.HandleProcessingError(ctx, operationID, metadata, "storage", attempts, err)
		}

		a.logger.Debug("Audio almacenado correctamente",
			zap.String("operation_id", operationID),
			zap.String("file_path", fileData.FilePath))

		if err := a.opsManager.HandleOperationSuccess(ctx, operationID, metadata, fileData); err != nil {
			a.logger.Error("Error al actualizar operación",
				zap.String("operation_id", operationID),
				zap.Error(err))
			return a.errorHandler.HandleProcessingError(ctx, operationID, metadata, "operation_update", attempts, err)
		}

		if err := a.messagingService.SendProcessingMessage(
			ctx, operationID, statusSuccess, metadata, attempts); err != nil {
			a.logger.Error("Error al enviar mensaje de procesamiento",
				zap.String("operation_id", operationID),
				zap.Error(err))
			return a.errorHandler.HandleProcessingError(ctx, operationID, metadata, "messaging", attempts, err)
		}

		a.logger.Info("Procesamiento completado exitosamente",
			zap.String("operation_id", operationID),
			zap.String("title", metadata.Title),
			zap.Int("attempts", attempts))

		return nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = a.config.Service.Timeout
	bo.MaxInterval = maxBackoff

	err := backoff.Retry(operation, bo)
	if err != nil {
		a.logger.Error("Procesamiento fallido después de reintentos",
			zap.String("operation_id", operationID),
			zap.Error(err),
			zap.Int("final_attempts", attempts))
	}
	return err
}

// createMetadata genera metadatos a partir de los detalles de un video de YouTube.
func (a *AudioProcessingService) createMetadata(mediaDetails *model.MediaDetails) *model.Metadata {
	return &model.Metadata{
		ID:           uuid.New().String(),
		VideoID:      mediaDetails.ID,
		Title:        mediaDetails.Title,
		DurationMs:   mediaDetails.DurationMs,
		URL:          mediaDetails.URL,
		Platform:     platformYoutube,
		ThumbnailURL: mediaDetails.Thumbnail,
	}
}
