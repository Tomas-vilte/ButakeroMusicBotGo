package player

// PlayerEvent representa eventos que pueden ocurrir en el reproductor
type PlayerEvent struct {
	Type    string
	Payload interface{}
}

// Event types
const (
	EventPlay   = "play"
	EventPause  = "pause"
	EventResume = "resume"
	EventStop   = "stop"
	EventSkip   = "skip"
)

// EventPayload contiene datos adicionales para los eventos
type EventPayload struct {
	TextChannelID  *string
	VoiceChannelID *string
}
