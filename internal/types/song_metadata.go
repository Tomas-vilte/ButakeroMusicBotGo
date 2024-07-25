package types

import "time"

type SongMetadata struct {
	Type         string        `json:"Type"`
	Title        string        `json:"Title"`
	URL          string        `json:"URL"`
	Playable     bool          `json:"Playable"`
	ThumbnailURL string        `json:"ThumbnailURL"`
	Duration     time.Duration `json:"Duration"`
}
