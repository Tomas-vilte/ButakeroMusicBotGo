package worker

import (
	"context"
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/ports"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"sync"
)

type WorkerPool struct {
	numWorkers int
	consumer   ports.MessageConsumer
	processor  Processor
	logger     logger.Logger
	workers    []*Worker
}

func NewWorkerPool(
	numWorkers int,
	consumer ports.MessageConsumer,
	processor Processor,
	logger logger.Logger,
) *WorkerPool {
	return &WorkerPool{
		numWorkers: numWorkers,
		consumer:   consumer,
		processor:  processor,
		logger:     logger,
		workers:    make([]*Worker, numWorkers),
	}
}

func (wp *WorkerPool) Start(ctx context.Context) error {
	requestChan, err := wp.consumer.GetRequestsChannel(ctx)
	if err != nil {
		return fmt.Errorf("error al obtener el canal de solicitudes: %w", err)
	}

	wp.logger.Info("Iniciando WorkerPool", zap.Int("num_workers", wp.numWorkers))

	var wg sync.WaitGroup

	for i := 0; i < wp.numWorkers; i++ {
		wg.Add(1)
		worker := NewWorker(i, wp.processor, wp.logger)
		wp.workers[i] = worker

		go worker.Start(ctx, &wg, requestChan)
	}

	<-ctx.Done()
	wp.logger.Info("Señal de cierre recibida, esperando a que los workers terminen")

	wg.Wait()
	wp.logger.Info("Todos los workers han terminado")

	return nil
}
