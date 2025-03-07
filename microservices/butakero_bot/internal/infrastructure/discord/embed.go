package discord

import (
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/bwmarrin/discordgo"
	"time"
)

// GeneratePlayingSongEmbed genera un embed para mostrar una canción en reproducción.
func GeneratePlayingSongEmbed(playMsg *entity.PlayedSong) *discordgo.MessageEmbed {
	if playMsg == nil || playMsg.Song == (entity.Song{}) {
		return nil
	}

	// Duración total en milisegundos
	durationMs := playMsg.Song.Metadata.DurationMs
	duration := time.Duration(durationMs) * time.Millisecond

	// Tiempo transcurrido en milisegundos
	elapsedMs := playMsg.Position
	elapsed := time.Duration(elapsedMs) * time.Millisecond

	// Barra de progreso
	progressBar := generateProgressBar(
		float64(elapsedMs)/float64(durationMs),
		20,
	)

	// Crear el embed
	embed := &discordgo.MessageEmbed{
		Title:       "🎵 **Reproduciendo:** " + playMsg.Song.Metadata.Title,
		Description: fmt.Sprintf("%s\n**%s / %s**", progressBar, formatDuration(elapsed), formatDuration(duration)),
		Color:       0x1DB954, // Color verde de Spotify
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "**Plataforma**",
				Value:  playMsg.Song.Metadata.Platform,
				Inline: true,
			},
			{
				Name:   "**Solicitado por**",
				Value:  playMsg.RequestedBy,
				Inline: true,
			},
		},
	}

	// Añadir la miniatura de la canción si está disponible
	if playMsg.Song.Metadata.ThumbnailURL != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: playMsg.Song.Metadata.ThumbnailURL,
		}
	}

	// Añadir un footer con información adicional
	embed.Footer = &discordgo.MessageEmbedFooter{
		Text: "Butakero Music Bot 🎶",
	}

	return embed
}

// formatDuration formatea una duración en formato MM:SS.
func formatDuration(duration time.Duration) string {
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
}

// generateProgressBar genera una barra de progreso visual.
func generateProgressBar(progress float64, length int) string {
	filled := int(progress * float64(length))
	bar := ""
	for i := 0; i < filled; i++ {
		bar += "▬"
	}
	bar += "🔘"
	for i := 0; i < length-filled-1; i++ {
		bar += "▬"
	}
	return bar
}
