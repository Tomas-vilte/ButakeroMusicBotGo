package player

import "context"

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
		HandleEvent(ctx context.Context, player *GuildPlayer) error
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

func (e PlayEvent) HandleEvent(ctx context.Context, player *GuildPlayer) error {
	return player.handlePlayEvent(ctx, e)
}

func (e PauseEvent) Type() EventType { return EventTypePause }

func (e PauseEvent) HandleEvent(ctx context.Context, player *GuildPlayer) error {
	return player.Pause(ctx)
}

func (e ResumeEvent) Type() EventType { return EventTypeResume }

func (e ResumeEvent) HandleEvent(ctx context.Context, player *GuildPlayer) error {
	return player.Resume(ctx)
}

func (e StopEvent) Type() EventType { return EventTypeStop }

func (e StopEvent) HandleEvent(ctx context.Context, player *GuildPlayer) error {
	return player.Stop(ctx)
}

func (e SkipEvent) Type() EventType { return EventTypeSkip }

func (e SkipEvent) HandleEvent(ctx context.Context, player *GuildPlayer) error {
	err := player.SkipSong(ctx)
	if err != nil {
		return err
	}
	return nil
}
