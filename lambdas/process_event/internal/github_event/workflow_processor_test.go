package github_event

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test con un evento válido de tipo WorkflowEvent
func TestProcessEvent_ValidWorkflowEvent(t *testing.T) {
	processor := &WorkflowEventProcessor{}

	workflowEvent := common.WorkflowEvent{
		Action: "completed",
		WorkFlowJobs: common.WorkFlowJob{
			WorkFlowName: "CI",
			ID:           1234,
		},
	}

	err := processor.ProcessEvent(context.Background(), workflowEvent)
	assert.NoError(t, err)
}

// Test con un evento inválido que no es de tipo WorkflowEvent
func TestProcessEvent_InvalidWorkflowEvent(t *testing.T) {
	processor := &WorkflowEventProcessor{}

	invalidEvent := struct {
		Action string
	}{Action: "invalid"}

	err := processor.ProcessEvent(context.Background(), invalidEvent)
	assert.Error(t, err)
	assert.Equal(t, "evento no es de tipo WorkflowEvent", err.Error())
}
