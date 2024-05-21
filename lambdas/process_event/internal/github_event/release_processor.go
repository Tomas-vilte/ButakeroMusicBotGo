package github_event

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/common"
)

type ReleaseEventProcessor struct{}

// ProcessEvent procesa un evento de lanzamiento.
func (r *ReleaseEventProcessor) ProcessEvent(ctx context.Context, event interface{}) error {
	releaseEvent, ok := event.(common.ReleaseEvent)
	if !ok {
		return errors.New("evento no es de tipo ReleaseEvent")
	}
	fmt.Printf("Procesando evento de release: TagName=%s, Name=%s\n", releaseEvent.Release.TagName, releaseEvent.Release.Name)
	return nil
}
