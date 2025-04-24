package entity

import "time"

type (
	DiscordEntity struct {
		ID           string
		TitleTrack   string
		DurationMs   int64
		ThumbnailURL string
		Platform     string
		FilePath     string
		URL          string
		AddedAt      time.Time
	}

	PlayedSong struct {
		DiscordSong     *DiscordEntity
		Position        int64
		RequestedByName string
		RequestedByID   string
		StartPosition   int64
	}
)
