package worker

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"go.uber.org/zap"
	"sync"
)

type Worker struct {
	id        int
	processor Processor
	logger    logger.Logger
}

func NewWorker(id int, processor Processor, logger logger.Logger) *Worker {
	return &Worker{
		id:        id,
		processor: processor,
		logger:    logger,
	}
}

func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup, requestChan <-chan *model.MediaRequest) {
	defer wg.Done()

	w.logger.Info("Iniciando worker", zap.Int("worker_id", w.id))

	for {
		select {
		case req, ok := <-requestChan:
			if !ok {
				w.logger.Info("Canal de solicitudes cerrado, finalizando worker", zap.Int("worker_id", w.id))
				return
			}

			w.logger.Info("Procesando requests", zap.Int("worker_id", w.id), zap.String("requests_id", req.RequestID))

			if err := w.processor.ProcessRequest(ctx, req); err != nil {
				w.logger.Error("Error al procesar la solicitud", zap.Int("worker_id", w.id), zap.String("requests_id", req.RequestID), zap.Error(err))
			}
		case <-ctx.Done():
			w.logger.Info("Cerrando senal, finalizando worker", zap.Int("worker_id", w.id))
			return
		}
	}

}
