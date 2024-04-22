package bot

import "time"

type PlayMessage struct {
	Song     *Song
	Position time.Duration
}

type Song struct {
	Type          string
	Title         string
	URL           string
	Playable      bool
	ThumbnailURL  *string
	Duration      time.Duration
	StartPosition time.Duration
	RequestedBy   *string
}

func (s Song) GetHumanName() any {

}
