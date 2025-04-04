package player

type (
	// PlayerEvent es la interfaz que todos los eventos del reproductor deben implementar
	PlayerEvent interface {
		isPlayerEvent()
		Type() string
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

func (e PlayEvent) Type() string   { return "play" }
func (e PlayEvent) isPlayerEvent() {}

func (e PauseEvent) Type() string   { return "pause" }
func (e PauseEvent) isPlayerEvent() {}

func (e ResumeEvent) Type() string   { return "resume" }
func (e ResumeEvent) isPlayerEvent() {}

func (e StopEvent) Type() string   { return "stop" }
func (e StopEvent) isPlayerEvent() {}

func (e SkipEvent) Type() string   { return "skip" }
func (e SkipEvent) isPlayerEvent() {}
