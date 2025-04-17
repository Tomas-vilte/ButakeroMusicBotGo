package entity

type (
	DiscordEntity struct {
		TitleTrack   string
		DurationMs   int64
		ThumbnailURL string
		Platform     string
		FilePath     string
		URL          string
	}

	PlayedSong struct {
		DiscordSong     *DiscordEntity
		Position        int64
		RequestedByName string
		RequestedByID   string
		StartPosition   int64
	}
)
