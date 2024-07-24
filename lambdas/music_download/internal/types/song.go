package types

import "time"

type Song struct {
	Type         string
	Title        string
	URL          string
	Playable     bool
	ThumbnailURL *string
	Duration     time.Duration
}
