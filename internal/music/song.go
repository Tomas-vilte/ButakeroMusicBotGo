package music

import "time"

// Song representa una canción obtenida de un servicio de música.
type Song struct {
	Type          string
	Title         string
	URL           string
	Playable      bool
	ThumbnailURL  *string
	Duration      time.Duration
	StartPosition time.Duration
}
