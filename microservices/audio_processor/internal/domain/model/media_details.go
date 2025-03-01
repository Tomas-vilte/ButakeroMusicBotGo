package model

import "time"

type MediaDetails struct {
	Title       string
	ID          string
	Description string
	Creator     string
	Duration    string
	PublishedAt time.Time
	URL         string
	Thumbnail   string
	Provider    string
}
