package player

type (

	// PlayerEvent representa eventos que pueden ocurrir en el reproductor
	PlayerEvent struct {
		Type    string
		Payload EventPayload
	}

	// EventPayload contiene datos adicionales para los eventos
	EventPayload struct {
		TextChannelID  *string
		VoiceChannelID *string
	}
)

// Event types
const (
	EventPlay   = "play"
	EventPause  = "pause"
	EventResume = "resume"
	EventStop   = "stop"
	EventSkip   = "skip"
)
