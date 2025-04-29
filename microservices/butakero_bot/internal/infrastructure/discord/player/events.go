package player

type EventType string

const (
	EventTypePlay   EventType = "play"
	EventTypePause  EventType = "pause"
	EventTypeResume EventType = "resume"
	EventTypeStop   EventType = "stop"
	EventTypeSkip   EventType = "skip"
)

type (
	// PlayerEvent es la interfaz que todos los eventos del reproductor deben implementar
	PlayerEvent interface {
		Type() EventType
	}
)

type (
	PlayEvent struct {
		TextChannelID  *string
		VoiceChannelID *string
	}

	PauseEvent struct{}

	ResumeEvent struct{}

	StopEvent struct{}

	SkipEvent struct{}
)

func (e PlayEvent) Type() EventType { return EventTypePlay }

func (e PauseEvent) Type() EventType { return EventTypePause }

func (e ResumeEvent) Type() EventType { return EventTypeResume }

func (e StopEvent) Type() EventType { return EventTypeStop }

func (e SkipEvent) Type() EventType { return EventTypeSkip }
