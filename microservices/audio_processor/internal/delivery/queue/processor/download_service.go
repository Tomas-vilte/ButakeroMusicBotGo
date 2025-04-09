package processor

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/worker"
)

type DownloadService struct {
	workerPool *worker.WorkerPool
	logger     logger.Logger
}

func NewDownloadService(
	numWorkers int,
	consumer ports.MessageConsumer,
	mediaService ports.MediaService,
	videoService ports.VideoService,
	coreService ports.CoreService,
	operationService ports.OperationService,
	logger logger.Logger,
) *DownloadService {
	processor := NewMediaProcessor(
		mediaService,
		videoService,
		coreService,
		operationService,
		logger,
	)

	workerPool := worker.NewWorkerPool(
		numWorkers,
		consumer,
		processor,
		logger,
	)

	return &DownloadService{
		workerPool: workerPool,
		logger:     logger,
	}
}

func (s *DownloadService) Run(ctx context.Context) error {
	s.logger.Info("Inciando servicio de descarga")
	return s.workerPool.Start(ctx)
}
