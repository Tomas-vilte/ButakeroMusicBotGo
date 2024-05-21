package github_event

import (
	"context"
	"errors"
	"fmt"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/common"
)

type WorkflowEventProcessor struct{}

// ProcessEvent procesa un evento de los estados de los deploys.
func (w *WorkflowEventProcessor) ProcessEvent(ctx context.Context, event interface{}) error {
	workflowEvent, ok := event.(common.WorkflowEvent)
	if !ok {
		return errors.New("evento no es de tipo WorkflowEvent")
	}

	fmt.Printf("Procesando evento de workflow: WorkFlowName=%s, ID=%v\n", workflowEvent.WorkFlowJobs.WorkFlowName, workflowEvent.WorkFlowJobs.ID)

	return nil
}
