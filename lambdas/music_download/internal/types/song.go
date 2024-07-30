package types

import "time"

type Song struct {
	Type         string        `json:"type"`
	Title        string        `json:"title"`
	URL          string        `json:"url"`
	Playable     bool          `json:"playable"`
	ThumbnailURL *string       `json:"thumbnail_url"`
	Duration     time.Duration `json:"duration"`
	Description  string        `json:"description"`
	Category     string        `json:"category"`
	Channel      Channel       `json:"channel"`
	PublishedAt  time.Time     `json:"published_at"`
	Tags         []string      `json:"tags"`
}

type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}
