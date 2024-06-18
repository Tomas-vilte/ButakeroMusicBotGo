package cache

import "container/list"

type (
	// ListInterface define las operaciones necesarias de una lista.
	ListInterface interface {
		PushFront(v interface{}) *list.Element
		MoveToFront(e *list.Element)
		Remove(e *list.Element)
		Back() *list.Element
		Len() int
	}

	StandardList struct {
		List *list.List
	}
)

func (s *StandardList) PushFront(v interface{}) *list.Element {
	return s.List.PushFront(v)
}

func (s *StandardList) MoveToFront(e *list.Element) {
	s.List.MoveToFront(e)
}

func (s *StandardList) Remove(e *list.Element) {
	s.List.Remove(e)
}

func (s *StandardList) Back() *list.Element {
	return s.List.Back()
}

func (s *StandardList) Len() int {
	return s.List.Len()
}

// newList crea una nueva instancia de ListInterface utilizando container/list.
func newList() ListInterface {
	return &StandardList{
		List: list.New(),
	}
}
