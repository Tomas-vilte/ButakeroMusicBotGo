package queuing

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReleaseEventFormatter_FormatEvent(t *testing.T) {
	formatter := &ReleaseEventFormatter{}

	event := map[string]interface{}{
		"release": map[string]interface{}{
			"tag_name": "v1.2.3",
			"body":     "Esta es la nueva release",
			"html_url": "https://example.com/release",
		},
	}

	expectedTitle := "🎉 ¡Salio una nueva version kkkk: v1.2.3!"
	expectedDescription := "Esta es la nueva release"
	expectedURL := "https://example.com/release"

	embed, err := formatter.FormatEvent(event)
	assert.NoError(t, err)
	assert.Equal(t, expectedTitle, embed.Title)
	assert.Equal(t, expectedDescription, embed.Description)
	assert.Equal(t, expectedURL, embed.URL)
	assert.Equal(t, 0x5865F2, embed.Color)
	assert.Equal(t, "ButakeroMusicBotGo", embed.Author.Name)
	assert.Equal(t, "https://cdn.discordapp.com/attachments/1231503103279366207/1243293979471122453/github.png?ex=6650f33f&is=664fa1bf&hm=3ececa29784b9549657bd52bc18e375ffd2f840a97a95bcee0bca97d1445a01b&", embed.Thumbnail.URL)
	assert.Len(t, embed.Fields, 1)
	assert.Equal(t, "Detalles de la Release", embed.Fields[0].Name)

}
