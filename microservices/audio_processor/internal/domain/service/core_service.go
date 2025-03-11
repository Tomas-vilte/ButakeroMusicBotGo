package service

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/config"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
	"time"
)

type CoreService struct {
	mediaService         ports.MediaService
	audioStorageService  ports.AudioStorageService
	topicPublisher       ports.TopicPublisherService
	videoService         ports.VideoService
	audioDownloadService ports.AudioDownloadService
	logger               logger.Logger
	cfg                  *config.Config
}

func NewCoreService(
	mediaService ports.MediaService,
	audioStorageService ports.AudioStorageService,
	topicPublisher ports.TopicPublisherService,
	audioDownloadService ports.AudioDownloadService,
	logger logger.Logger,
	cfg *config.Config,
) *CoreService {
	return &CoreService{
		mediaService:         mediaService,
		audioStorageService:  audioStorageService,
		topicPublisher:       topicPublisher,
		audioDownloadService: audioDownloadService,
		logger:               logger,
		cfg:                  cfg,
	}
}

func (s *CoreService) ProcessMedia(ctx context.Context, operationID string, mediaDetails *model.MediaDetails) error {
	log := s.logger.With(
		zap.String("component", "CoreService"),
		zap.String("method", "ProcessMedia"),
		zap.String("song", mediaDetails.ID),
	)

	ctx, cancel := context.WithTimeout(ctx, s.cfg.Service.Timeout)
	defer cancel()

	metadata := s.createMetadata(mediaDetails)
	attempts := 0

	s.logger.Info("Iniciando procesamiento de audio",
		zap.String("operation_id", operationID),
		zap.String("title", metadata.Title),
	)

	operation := func() error {
		attempts++

		if attempts > s.cfg.Service.MaxAttempts {
			return backoff.Permanent(fmt.Errorf("número máximo de intentos alcanzado (%d)", s.cfg.Service.MaxAttempts))
		}

		if attempts > 1 {
			s.logger.Info("Reintentando procesamiento",
				zap.String("operation_id", operationID),
				zap.Int("attempt", attempts),
				zap.Int("max_attempts", s.cfg.Service.MaxAttempts))
		}

		audioBuffer, err := s.audioDownloadService.DownloadAndEncode(ctx, mediaDetails.URL)
		if err != nil {
			log.Error("Error al descargar y codificar el audio", zap.Error(err))
			return fmt.Errorf("error al descargar y codificar el audio: %w", err)
		}

		fileData, err := s.audioStorageService.StoreAudio(ctx, audioBuffer, mediaDetails.Title)
		if err != nil {
			log.Error("Error al almacenar el archivo de audio", zap.Error(err))
			return fmt.Errorf("error al almacenar el archivo de audio: %w", err)
		}

		media := &model.Media{
			ID:      operationID,
			VideoID: mediaDetails.ID,
			Status:  "success",
			Message: "Procesamiento completado exitosamente",
			Metadata: &model.PlatformMetadata{
				Title:        mediaDetails.Title,
				DurationMs:   mediaDetails.DurationMs,
				URL:          mediaDetails.URL,
				ThumbnailURL: mediaDetails.ThumbnailURL,
				Platform:     "youtube",
			},
			FileData:       fileData,
			ProcessingDate: time.Now(),
			Success:        true,
			Attempts:       1,
			Failures:       0,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		message := &model.MediaProcessingMessage{
			ID:               media.ID,
			VideoID:          media.VideoID,
			FileData:         media.FileData,
			PlatformMetadata: media.Metadata,
		}

		if err := s.mediaService.UpdateMedia(ctx, media.ID, media.VideoID, media); err != nil {
			log.Error("Error al actualizar operation", zap.String("operation_id", operationID), zap.Error(err))
			return fmt.Errorf("error al actualizar operation: %w", err)
		}

		if err := s.topicPublisher.PublishMediaProcessed(ctx, message); err != nil {
			log.Error("Error al publicar el evento de procesamiento exitoso", zap.Error(err))
			return fmt.Errorf("error al publicar el evento: %w", err)
		}

		log.Info("Procesamiento de medios completado exitosamente", zap.String("media_id", media.ID))
		return nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = s.cfg.Service.Timeout
	bo.MaxInterval = 30 * time.Second

	err := backoff.Retry(operation, bo)
	if err != nil {
		s.logger.Error("Procesamiento fallido después de reintentos",
			zap.String("operation_id", operationID),
			zap.Error(err),
			zap.Int("final_attempts", attempts))
		return err
	}

	return nil

}

func (s *CoreService) createMetadata(mediaDetails *model.MediaDetails) *model.PlatformMetadata {
	return &model.PlatformMetadata{
		Title:        mediaDetails.Title,
		DurationMs:   mediaDetails.DurationMs,
		URL:          mediaDetails.URL,
		ThumbnailURL: mediaDetails.ThumbnailURL,
		Platform:     mediaDetails.Provider,
	}
}
