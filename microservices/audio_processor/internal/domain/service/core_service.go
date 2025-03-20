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

type coreService struct {
	mediaService         ports.MediaService
	audioStorageService  ports.AudioStorageService
	topicPublisher       ports.TopicPublisherService
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
) ports.CoreService {
	return &coreService{
		mediaService:         mediaService,
		audioStorageService:  audioStorageService,
		topicPublisher:       topicPublisher,
		audioDownloadService: audioDownloadService,
		logger:               logger,
		cfg:                  cfg,
	}
}

func (s *coreService) ProcessMedia(ctx context.Context, mediaDetails *model.MediaDetails) error {
	log := s.logger.With(
		zap.String("component", "CoreService"),
		zap.String("method", "ProcessMedia"),
		zap.String("song", mediaDetails.ID),
	)

	ctx, cancel := context.WithTimeout(ctx, s.cfg.Service.Timeout)
	defer cancel()

	attempts := 0

	s.logger.Info("Iniciando procesamiento de audio",
		zap.String("title", mediaDetails.Title),
	)

	var lastError error

	operation := func() error {
		attempts++

		if attempts > s.cfg.Service.MaxAttempts {
			return backoff.Permanent(fmt.Errorf("número máximo de intentos alcanzado (%d): %w", s.cfg.Service.MaxAttempts, lastError))
		}

		if attempts > 1 {
			s.logger.Info("Reintentando procesamiento",
				zap.Int("attempt", attempts),
				zap.Int("max_attempts", s.cfg.Service.MaxAttempts))
		}

		audioBuffer, err := s.audioDownloadService.DownloadAndEncode(ctx, mediaDetails.URL)
		if err != nil {
			log.Error("Error al descargar y codificar el audio", zap.Error(err))
			lastError = err
			return lastError
		}

		fileData, err := s.audioStorageService.StoreAudio(ctx, audioBuffer, mediaDetails.Title)
		if err != nil {
			log.Error("Error al almacenar el archivo de audio", zap.Error(err))
			lastError = err
			return lastError
		}

		media := &model.Media{
			VideoID:    mediaDetails.ID,
			TitleLower: mediaDetails.Title,
			Status:     "success",
			Message:    "Procesamiento completado exitosamente",
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
			VideoID:          media.VideoID,
			FileData:         media.FileData,
			PlatformMetadata: media.Metadata,
			Status:           media.Status,
			Success:          media.Success,
			Message:          media.Message,
		}

		if err := s.mediaService.UpdateMedia(ctx, media.VideoID, media); err != nil {
			log.Error("Error al actualizar la operación", zap.String("video_id", media.VideoID), zap.Error(err))
			lastError = err
			return lastError
		}

		if err := s.topicPublisher.PublishMediaProcessed(ctx, message); err != nil {
			log.Error("Error al publicar el evento de procesamiento exitoso", zap.Error(err))
			lastError = err
			return lastError
		}

		log.Info("Procesamiento de medios completado exitosamente", zap.String("video_id", media.VideoID))
		return nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = s.cfg.Service.Timeout
	bo.MaxInterval = 30 * time.Second

	err := backoff.Retry(operation, bo)
	if err != nil {
		s.logger.Error("Procesamiento fallido después de reintentos",
			zap.Error(err),
			zap.Int("final_attempts", attempts))
		return err
	}

	return nil
}
