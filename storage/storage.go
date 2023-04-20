package storage

import (
	"sync"
)

func NewEventStorage[T comparable](size int) *eventStorage[T] {
	return &eventStorage[T]{
		size:   size,
		events: make(chan T, size),
		data: sync.Pool{New: func() any {
			data := make([]T, 0, size)
			return data
		}},
	}
}

type eventStorage[T comparable] struct {
	size   int
	events chan T
	data   sync.Pool
}

func (s *eventStorage[T]) Put(e T) bool {
	s.events <- e
	return len(s.events) < s.size
}

func (s *eventStorage[T]) Get() []T {
	dataPool := s.data.Get()
	data, _ := dataPool.([]T)

	l := len(s.events) // fix chan size
	for i := 0; i < l; i++ {
		data = append(data, <-s.events)
	}
	s.data.Put(dataPool)

	return data
}
