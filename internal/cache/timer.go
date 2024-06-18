package cache

import (
	"time"
)

type (
	// TimerInterface define las operaciones necesarias para un temporizador.
	TimerInterface interface {
		C() <-chan time.Time
		Reset(d time.Duration) bool
		Stop() bool
	}

	StandardTimer struct {
		Timer *time.Timer
	}
)

func (s *StandardTimer) C() <-chan time.Time {
	return s.Timer.C
}

func (s *StandardTimer) Reset(d time.Duration) bool {
	return s.Timer.Reset(d)
}

func (s *StandardTimer) Stop() bool {
	return s.Timer.Stop()
}

// newTimer crea una nueva instancia de TimerInterface utilizando time.Timer.
func newTimer(d time.Duration) TimerInterface {
	return &StandardTimer{
		Timer: time.NewTimer(d),
	}
}
