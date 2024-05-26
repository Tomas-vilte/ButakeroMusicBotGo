package queuing

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestWorkflowActionEventFormatter_FormatEvent(t *testing.T) {
	formatter := &WorkflowActionEventFormatter{}

	event := map[string]interface{}{
		"workflow_job": map[string]interface{}{
			"workflow_name": "CI Pipeline",
			"conclusion":    "success",
			"html_url":      "https://example.com/workflow",
			"completed_at":  "2021-01-01T12:00:00Z",
		},
		"action": "completed",
	}

	expectedTitle := "🚀 Acción completada en el flujo de trabajo"
	expectedDescription := "El flujo de trabajo **CI Pipeline** ha completado una acción:"
	expectedConclusion := "**success**"
	expectedState := "**completed**"

	embed, err := formatter.FormatEvent(event)
	assert.NoError(t, err)
	assert.Equal(t, expectedTitle, embed.Title)
	assert.Equal(t, expectedDescription, embed.Description)
	assert.Equal(t, 0x34a853, embed.Color)
	assert.Len(t, embed.Fields, 3)
	assert.Equal(t, "Estado", embed.Fields[0].Name)
	assert.Equal(t, expectedState, embed.Fields[0].Value)
	assert.Equal(t, "Conclusión", embed.Fields[1].Name)
	assert.Equal(t, expectedConclusion, embed.Fields[1].Value)
	assert.Equal(t, "Detalles de la acción", embed.Fields[2].Name)
	assert.Equal(t, "[Ver detalles](https://example.com/workflow)", embed.Fields[2].Value)
	assert.Equal(t, "2021-01-01T12:00:00Z", embed.Timestamp)
	assert.Equal(t, "https://cdn.discordapp.com/attachments/1231503103279366207/1243293979471122453/github.png?ex=6650f33f&is=664fa1bf&hm=3ececa29784b9549657bd52bc18e375ffd2f840a97a95bcee0bca97d1445a01b&", embed.Thumbnail.URL)
}
