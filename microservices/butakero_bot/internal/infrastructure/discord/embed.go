package discord

import (
	"fmt"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/entity"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/butakero_bot/internal/domain/model/discord"
	"time"
)

// GeneratePlayingSongEmbed genera un embed para mostrar una canción en reproducción.
func GeneratePlayingSongEmbed(playMsg *entity.PlayedSong) *discord.Embed {
	if playMsg == nil || playMsg.DiscordSong == nil {
		return nil
	}

	durationMs := playMsg.DiscordSong.DurationMs
	duration := time.Duration(durationMs) * time.Millisecond

	elapsedMs := playMsg.Position
	elapsed := time.Duration(elapsedMs) * time.Millisecond

	progressBar := generateProgressBar(
		float64(elapsedMs)/float64(durationMs),
		20,
	)

	embed := &discord.Embed{
		Title:       "🎵 **Reproduciendo:** " + playMsg.DiscordSong.TitleTrack,
		Description: fmt.Sprintf("%s\n**%s / %s**", progressBar, formatDuration(elapsed), formatDuration(duration)),
		Color:       0x1DB954,
		Fields: []*discord.EmbedField{
			{
				Name:   "**Plataforma**",
				Value:  playMsg.DiscordSong.Platform,
				Inline: true,
			},
			{
				Name:   "**Solicitado por**",
				Value:  playMsg.RequestedByName,
				Inline: true,
			},
		},
	}

	if playMsg.DiscordSong.ThumbnailURL != "" {
		embed.Thumbnail = &discord.EmbedThumbnail{
			URL:    playMsg.DiscordSong.ThumbnailURL,
			Width:  1280,
			Height: 720,
		}
	}

	embed.Footer = &discord.EmbedFooter{
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
