package worker

import (
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/domain/model"
)

type Processor interface {
	ProcessRequest(ctx context.Context, req *model.MediaRequest) error
}
