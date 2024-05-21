package github_event

import (
	"context"
	"github.com/Tomas-vilte/GoMusicBot/lambdas/process_event/internal/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test con un evento válido de tipo ReleaseEvent
func TestProcessEvent_ValidReleaseEvent(t *testing.T) {
	processor := &ReleaseEventProcessor{}

	releaseEvent := common.ReleaseEvent{
		Action: "published",
		Release: common.Release{
			TagName: "v1.0.0",
			Name:    "Initial Release",
		},
	}

	err := processor.ProcessEvent(context.Background(), releaseEvent)
	assert.NoError(t, err)
}

// Test con un evento inválido que no es de tipo ReleaseEvent
func TestProcessEvent_InvalidEvent(t *testing.T) {
	processor := &ReleaseEventProcessor{}

	invalidEvent := struct {
		Action string
	}{Action: "invalid"}

	err := processor.ProcessEvent(context.Background(), invalidEvent)
	assert.Error(t, err)
	assert.Equal(t, "evento no es de tipo ReleaseEvent", err.Error())
}
