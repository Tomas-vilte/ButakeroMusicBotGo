package processor

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/worker"
	"go.uber.org/zap"
)

type DownloadService struct {
	workerPool worker.DownloadSongWorkerPool
	logger     logger.Logger
}

func NewDownloadService(workerPool worker.DownloadSongWorkerPool, logger logger.Logger) *DownloadService {
	return &DownloadService{
		workerPool: workerPool,
		logger:     logger,
	}
}

func (s *DownloadService) Run(ctx context.Context) error {
	log := s.logger.With(
		zap.String("component", "DownloadService"),
		zap.String("method", "Run"),
	)
	log.Info("Inciando servicio de descarga")
	if err := s.workerPool.Start(ctx); err != nil {
		log.Error("Error al iniciar el servicio de descarga", zap.Error(err))
		return err
	}
	return nil
}
