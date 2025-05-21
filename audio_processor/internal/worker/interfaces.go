package worker

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"sync"
)

type (
	AudioTaskProcessor interface {
		ProcessDownloadTask(ctx context.Context, req *model.MediaRequest) error
	}

	TaskWorker interface {
		Run(ctx context.Context, wg *sync.WaitGroup, taskChan <-chan *model.MediaRequest)
	}

	WorkerFactory interface {
		NewWorker(id int, processor AudioTaskProcessor, logger logger.Logger) TaskWorker
	}

	DownloadSongWorkerPool interface {
		Start(ctx context.Context) error
	}
)
