package voice

import (
	"github.com/Tomas-vilte/GoMusicBot/internal/utils"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGeneratePlayingSongEmbed_ValidMessage(t *testing.T) {
	// Configuración
	message := &PlayMessage{
		Song: &Song{
			Title:        "Canción de prueba",
			Duration:     180 * time.Second, // 3 minutos
			ThumbnailURL: utils.String("https://ejemplo.com/imagen.png"),
			RequestedBy:  utils.String("Usuario de prueba"),
		},
		Position: 120 * time.Second, // 2 minutos de reproducción
	}

	// Ejecución
	embed := GeneratePlayingSongEmbed(message)

	// Verificación
	assert.NotNil(t, embed)
	assert.Equal(t, "Canción de prueba", embed.Title)
	assert.Contains(t, embed.Description, "▬🔘")            // Verifica que haya una barra de progreso
	assert.Contains(t, embed.Description, "02:00 / 03:00") // Verifica la duración
	assert.NotNil(t, embed.Thumbnail)
	assert.Equal(t, "https://ejemplo.com/imagen.png", embed.Thumbnail.URL)
	assert.NotNil(t, embed.Footer)
	assert.Equal(t, "Solicitado por: Usuario de prueba", embed.Footer.Text)
}

func TestGeneratePlayingSongEmbed_NilMessage(t *testing.T) {
	// Configuración
	var message *PlayMessage

	// Ejecución
	embed := GeneratePlayingSongEmbed(message)

	// Verificación
	assert.Nil(t, embed)
}

func TestGeneratePlayingSongEmbed_NilSong(t *testing.T) {
	// Configuración
	message := &PlayMessage{} // No se define un Song

	// Ejecución
	embed := GeneratePlayingSongEmbed(message)

	// Verificación
	assert.Nil(t, embed)
}

func TestGeneratePlayingSongEmbed_NilThumbnailURL(t *testing.T) {
	// Configuración
	message := &PlayMessage{
		Song: &Song{
			Title:       "Canción de prueba",
			Duration:    180,
			RequestedBy: utils.String("Usuario de prueba"),
		},
		Position: 120,
	}

	// Ejecución
	embed := GeneratePlayingSongEmbed(message)

	// Verificación
	assert.NotNil(t, embed)
	assert.Nil(t, embed.Thumbnail)
}

func TestGeneratePlayingSongEmbed_NilRequestedBy(t *testing.T) {
	// Configuración
	message := &PlayMessage{
		Song: &Song{
			Title:        "Canción de prueba",
			Duration:     180,
			ThumbnailURL: utils.String("https://ejemplo.com/imagen.png"),
		},
		Position: 120,
	}

	// Ejecución
	embed := GeneratePlayingSongEmbed(message)

	// Verificación
	assert.NotNil(t, embed)
	assert.Nil(t, embed.Footer)
}
