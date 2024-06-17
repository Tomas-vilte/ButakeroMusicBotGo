package cache

import "sync"

type (
	// EntryPoolInterface define las operaciones necesarias para un pool de entradas.
	EntryPoolInterface interface {
		Get() *Entry
		Put(e *Entry)
	}

	StandardEntryPool struct {
		Pool sync.Pool
	}
)

func (s *StandardEntryPool) Get() *Entry {
	return s.Pool.Get().(*Entry)
}

func (s *StandardEntryPool) Put(e *Entry) {
	s.Pool.Put(e)
}

// newEntryPool crea una nueva instancia de EntryPoolInterface utilizando sync.Pool.
func newEntryPool() EntryPoolInterface {
	return &StandardEntryPool{
		Pool: sync.Pool{
			New: func() interface{} {
				return &Entry{}
			},
		},
	}
}

type (
	// EntryPoolInterfaceAudio define las operaciones necesarias para un pool de entradas.
	EntryPoolInterfaceAudio interface {
		Get() *EntryAudioCaching
		Put(e *EntryAudioCaching)
	}

	StandardEntryPoolAudio struct {
		Pool sync.Pool
	}
)

func (s *StandardEntryPoolAudio) Get() *EntryAudioCaching {
	return s.Pool.Get().(*EntryAudioCaching)
}

func (s *StandardEntryPoolAudio) Put(e *EntryAudioCaching) {
	s.Pool.Put(e)
}

// newEntryPoolAudio crea una nueva instancia de EntryPoolInterfaceAudio utilizando sync.Pool.
func newEntryPoolAudio() EntryPoolInterfaceAudio {
	return &StandardEntryPoolAudio{
		Pool: sync.Pool{
			New: func() interface{} {
				return &EntryAudioCaching{}
			},
		},
	}
}
