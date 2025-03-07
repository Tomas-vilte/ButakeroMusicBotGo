package model

import "time"

type MediaDetails struct {
	Title       string
	ID          string
	Description string
	Creator     string
	DurationMs  int64
	PublishedAt time.Time
	URL         string
	Thumbnail   string
	Provider    string
}
