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
	mediaRepository      ports.MediaRepository
	audioStorageService  ports.AudioStorageService
	topicPublisher       ports.MessageProducer
	audioDownloadService ports.AudioDownloadService
	logger               logger.Logger
	cfg                  *config.Config
}

func NewCoreService(
	mediaRepository ports.MediaRepository,
	audioStorageService ports.AudioStorageService,
	topicPublisher ports.MessageProducer,
	audioDownloadService ports.AudioDownloadService,
	logger logger.Logger,
	cfg *config.Config,
) ports.CoreService {
	return &coreService{
		mediaRepository:      mediaRepository,
		audioStorageService:  audioStorageService,
		topicPublisher:       topicPublisher,
		audioDownloadService: audioDownloadService,
		logger:               logger,
		cfg:                  cfg,
	}
}

func (s *coreService) ProcessMedia(ctx context.Context, media *model.Media, userID, requestID string) error {
	log := s.logger.With(
		zap.String("component", "CoreService"),
		zap.String("method", "ProcessMedia"),
		zap.String("song", media.VideoID),
	)

	if err := media.Validate(); err != nil {
		return fmt.Errorf("error al validar el media: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, s.cfg.Service.Timeout)
	defer cancel()

	attempts := 0

	s.logger.Info("Iniciando procesamiento de audio",
		zap.String("title", media.Metadata.Title),
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

		audioBuffer, err := s.audioDownloadService.DownloadAndEncode(ctx, media.Metadata.URL)
		if err != nil {
			log.Error("Error al descargar y codificar el audio", zap.Error(err))
			lastError = err
			return err
		}

		fileData, err := s.audioStorageService.StoreAudio(ctx, audioBuffer, media.TitleLower)
		if err != nil {
			log.Error("Error al almacenar el archivo de audio", zap.Error(err))
			lastError = err
			return err
		}

		media.UpdateAsSuccess(fileData, attempts)
		if err := s.mediaRepository.UpdateMedia(ctx, media.VideoID, media); err != nil {
			log.Error("Error al actualizar la operación", zap.String("video_id", media.VideoID), zap.Error(err))
			lastError = err
			return err
		}

		if err := s.topicPublisher.Publish(ctx, media.ToMessage(requestID, userID)); err != nil {
			log.Error("Error al publicar el evento de procesamiento exitoso", zap.Error(err))
			lastError = err
			return err
		}

		log.Info("Procesamiento de medios completado exitosamente", zap.String("video_id", media.VideoID))
		return nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = s.cfg.Service.Timeout
	return backoff.RetryNotify(operation, bo, func(err error, d time.Duration) {
		log.Warn("Reintentando después de error", zap.Error(err), zap.Duration("delay", d))
	})
}
