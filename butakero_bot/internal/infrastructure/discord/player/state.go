package player

import (
	"sync"
)

// StateManager maneja el estado del reproductor
type StateManager struct {
	mu    sync.Mutex
	state PlayerState
}

func NewStateManager() *StateManager {
	return &StateManager{state: StateIdle}
}

func (sm *StateManager) SetState(newState PlayerState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.state = newState
}

func (sm *StateManager) GetState() PlayerState {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	return sm.state
}
