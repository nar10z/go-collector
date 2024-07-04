/*
 * Copyright (c) 2023.
 *
 * License MIT (https://raw.githubusercontent.com/nar10z/go-accumulator/main/LICENSE)
 *
 * Developed thanks to Nikita Terentyev (nar10z). Use it for good, and let your code work without problems!
 */

package goaccum

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultFlushSize     = 1000
	defaultFlushInterval = time.Millisecond * 250
)

// New creates a new data Accumulator
func New[T any](
	flushSize uint,
	flushInterval time.Duration,
	flushFunc FlushExec[T],
) *Accumulator[T] {
	if flushSize == 0 {
		flushSize = defaultFlushSize
	}

	if flushInterval == 0 {
		flushInterval = defaultFlushInterval
	}

	if flushFunc == nil {
		flushFunc = noop[T]
	}

	a := &Accumulator[T]{
		flushFunc: flushFunc,

		chEvents: make(chan eventExtended[T], flushSize),
		batchEvents: sync.Pool{
			New: func() any {
				return make([]eventExtended[T], 0, flushSize)
			},
		},
		batchOrigEvents: sync.Pool{
			New: func() any {
				return make([]T, 0, flushSize)
			},
		},

		chStop: make(chan struct{}),
	}

	go a.startFlusher(flushInterval, int(flushSize))

	return a
}

type Accumulator[T any] struct {
	batchEvents     sync.Pool
	batchOrigEvents sync.Pool
	flushFunc       FlushExec[T]
	chEvents        chan eventExtended[T]
	chStop          chan struct{}
	isClose         atomic.Bool
}

func (a *Accumulator[T]) AddAsync(ctx context.Context, event T) error {
	if a.isClose.Load() {
		return ErrSendToClose
	}

	select {
	case <-ctx.Done():
		return fmt.Errorf("AddAsync, check on write: %w", ctx.Err())
	default:
		a.chEvents <- eventExtended[T]{e: event}
	}

	return nil
}

func (a *Accumulator[T]) AddSync(ctx context.Context, event T) error {
	// check context before alloc eventExtended
	select {
	case <-ctx.Done():
		return fmt.Errorf("AddSync, check before: %w", ctx.Err())
	default:
	}

	e := eventExtended[T]{
		fallback: make(chan error),
		e:        event,
	}

	if a.isClose.Load() {
		return ErrSendToClose
	}

	// check context with write to channel
	select {
	case <-ctx.Done():
		return fmt.Errorf("AddSync, check on write: %w", ctx.Err())
	case a.chEvents <- e:
	}

	// check context with wait event result
	select {
	case err := <-e.fallback:
		if err != nil {
			return fmt.Errorf("AddSync, check fallback: %w", err)
		}

		return nil
	case <-ctx.Done():
		e.fallback = nil
		return fmt.Errorf("AddSync, check fallback: %w", ctx.Err())
	}
}

func (a *Accumulator[T]) Stop() {
	if !a.isClose.CompareAndSwap(false, true) {
		return
	}

	close(a.chEvents)
	<-a.chStop
}

func (a *Accumulator[T]) IsClosed() bool {
	return a.isClose.Load()
}

func (a *Accumulator[T]) startFlusher(interval time.Duration, size int) {
	ticker := time.NewTicker(interval)
	batch, _ := a.batchEvents.Get().([]eventExtended[T])
	flush := func() {
		a.flush(batch)
		a.batchEvents.Put(batch[:0])
		batch, _ = a.batchEvents.Get().([]eventExtended[T])
	}

loop:
	for {
		select {
		case e, ok := <-a.chEvents:
			if !ok {
				break loop
			}

			batch = append(batch, e)
			if len(batch) < size {
				continue
			}

			flush()
		case <-ticker.C:
			flush()
		}
	}

	ticker.Stop()
	a.chEvents = nil
	flush()
	a.chStop <- struct{}{}
}

func (a *Accumulator[T]) flush(events []eventExtended[T]) {
	if len(events) == 0 {
		return
	}

	originalEvents, _ := a.batchOrigEvents.Get().([]T)
	for i := 0; i < len(events); i++ {
		originalEvents = append(originalEvents, events[i].e)
	}

	err := a.flushFunc(originalEvents)
	for i := 0; i < len(events); i++ {
		if events[i].fallback == nil {
			continue
		}

		events[i].fallback <- err
	}

	a.batchOrigEvents.Put(originalEvents[:0])
}
