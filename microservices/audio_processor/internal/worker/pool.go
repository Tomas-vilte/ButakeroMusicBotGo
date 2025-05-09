package worker

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"sync"
)

type DownloadWorkerPool struct {
	workerCount   int
	consumer      ports.MessageConsumer
	processor     AudioTaskProcessor
	logger        logger.Logger
	workerFactory WorkerFactory
}

func NewDownloadWorkerPool(
	workerCount int,
	consumer ports.MessageConsumer,
	processor AudioTaskProcessor,
	logger logger.Logger,
	workerFactory WorkerFactory,
) *DownloadWorkerPool {
	return &DownloadWorkerPool{
		workerCount:   workerCount,
		consumer:      consumer,
		processor:     processor,
		logger:        logger,
		workerFactory: workerFactory,
	}
}

func (wp *DownloadWorkerPool) Start(ctx context.Context) error {
	taskChan, err := wp.consumer.GetRequestsChannel(ctx)
	if err != nil {
		return fmt.Errorf("error al obtener el canal de solicitudes: %w", err)
	}

	wp.logger.Info("Iniciando DownloadWorkerPool", zap.Int("num_workers", wp.workerCount))

	var wg sync.WaitGroup

	for i := 0; i < wp.workerCount; i++ {
		wg.Add(1)
		worker := wp.workerFactory.NewWorker(i, wp.processor, wp.logger)
		go worker.Run(ctx, &wg, taskChan)
	}

	<-ctx.Done()
	wp.logger.Info("Señal de cierre recibida, esperando a que los workers terminen")

	wg.Wait()
	wp.logger.Info("Todos los workers han terminado")

	return nil
}
