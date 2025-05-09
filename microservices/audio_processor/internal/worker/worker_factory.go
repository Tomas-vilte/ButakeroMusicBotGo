package worker

import "github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"

type workerFactory struct{}

func (wf *workerFactory) NewWorker(id int, processor AudioTaskProcessor, logger logger.Logger) TaskWorker {
	return NewDownloadTaskWorker(id, processor, logger)
}

func NewWorkerFactory() WorkerFactory {
	return &workerFactory{}
}
