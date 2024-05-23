package github_event

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/common"
)

type ReleaseEventProcessor struct{}

// ProcessEvent procesa un evento de lanzamiento.
func (r *ReleaseEventProcessor) ProcessEvent(ctx context.Context, event interface{}, eventType string) error {
	// Verificar si el evento es del tipo correcto
	if eventType != "release" {
		return errors.New("evento no es de tipo 'release'")
	}

	releaseEvent, ok := event.(common.ReleaseEvent)
	if !ok {
		return errors.New("evento no es de tipo ReleaseEvent")
	}
	fmt.Printf("Procesando evento de release: TagName=%s, Name=%s\n", releaseEvent.Release.TagName, releaseEvent.Release.Name)
	return nil
}
