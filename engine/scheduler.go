package engine

import (
	"log"
	"sync"
	"time"
)

// Scheduler coalesces update requests into at most one flush per tick. Callers
// mark the tree dirty via Enqueue; on each interval (typically 16ms for 60fps)
// the scheduler runs the registered flush function once if anything changed,
// collapsing multiple updates within a frame into a single render.
//
// Updates have two priorities. Enqueue is the default: it waits for the next
// tick, so a burst of low-stakes mutations costs a single flush per frame.
// EnqueueNow is for latency-sensitive updates (e.g. reflecting a keystroke or a
// click) — it wakes the loop immediately so the change reaches the client
// without waiting up to a full frame. Both priorities still coalesce: many
// EnqueueNow calls in the same instant collapse into one flush.
type Scheduler struct {
	interval time.Duration
	mu       sync.Mutex
	dirty    bool
	flushFn  func()
	ticker   *time.Ticker
	wake     chan struct{}
	done     chan struct{}
}

// NewScheduler returns a Scheduler that ticks every interval once started.
func NewScheduler(interval time.Duration) *Scheduler {
	return &Scheduler{
		interval: interval,
		wake:     make(chan struct{}, 1),
		done:     make(chan struct{}),
	}
}

// SetFlush registers the function invoked once per tick when the scheduler is
// dirty. It must be called before Start.
func (s *Scheduler) SetFlush(fn func()) {
	s.flushFn = fn
}

// Enqueue marks the scheduler dirty so the next tick triggers a flush. Multiple
// calls within one interval coalesce into a single flush.
func (s *Scheduler) Enqueue() {
	s.mu.Lock()
	s.dirty = true
	s.mu.Unlock()
}

// EnqueueNow marks the scheduler dirty and wakes the loop immediately rather
// than waiting for the next tick, trading a little extra work for lower latency
// on updates that must feel instant. It still coalesces: if a flush is already
// pending this frame, the wake is a no-op.
func (s *Scheduler) EnqueueNow() {
	s.mu.Lock()
	s.dirty = true
	s.mu.Unlock()

	// Non-blocking: the buffered channel holds at most one pending wake, so a
	// burst of EnqueueNow calls collapses into a single extra flush.
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

// Start begins the ticker and runs the flush loop in a background goroutine.
func (s *Scheduler) Start() {
	s.ticker = time.NewTicker(s.interval)
	go s.loop()
}

func (s *Scheduler) loop() {
	for {
		select {
		case <-s.ticker.C:
			s.maybeFlush()
		case <-s.wake:
			s.maybeFlush()
		case <-s.done:
			return
		}
	}
}

// maybeFlush runs the flush function once if the scheduler is dirty, clearing
// the dirty flag first so concurrent Enqueue calls during the flush schedule a
// fresh flush rather than being lost.
func (s *Scheduler) maybeFlush() {
	s.mu.Lock()
	shouldFlush := s.dirty && s.flushFn != nil
	s.dirty = false
	s.mu.Unlock()

	if shouldFlush {
		s.flushFn()
	}
}

// Stop halts the flush loop and stops the underlying ticker.
func (s *Scheduler) Stop() {
	log.Println("[scheduler] stopping")
	close(s.done)
	if s.ticker != nil {
		s.ticker.Stop()
	}
}
