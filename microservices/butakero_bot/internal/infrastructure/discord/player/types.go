package player

// PlayerState representa el estado actual del reproductor
type PlayerState string

const (
	StateIdle    PlayerState = "idle"
	StatePlaying PlayerState = "playing"
	StatePaused  PlayerState = "paused"
)
