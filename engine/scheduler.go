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
type Scheduler struct {
	interval time.Duration
	mu       sync.Mutex
	dirty    bool
	flushFn  func()
	ticker   *time.Ticker
	done     chan struct{}
}

// NewScheduler returns a Scheduler that ticks every interval once started.
func NewScheduler(interval time.Duration) *Scheduler {
	return &Scheduler{
		interval: interval,
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

// Start begins the ticker and runs the flush loop in a background goroutine.
func (s *Scheduler) Start() {
	s.ticker = time.NewTicker(s.interval)
	go s.loop()
}

func (s *Scheduler) loop() {
	for {
		select {
		case <-s.ticker.C:
			s.mu.Lock()
			shouldFlush := s.dirty && s.flushFn != nil
			s.dirty = false
			s.mu.Unlock()

			if shouldFlush {
				s.flushFn()
			}
		case <-s.done:
			return
		}
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
