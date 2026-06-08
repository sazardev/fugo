package engine

import (
	"log"
	"sync"
	"time"
)

type Scheduler struct {
	interval time.Duration
	mu       sync.Mutex
	dirty    bool
	flushFn  func()
	ticker   *time.Ticker
	done     chan struct{}
}

func NewScheduler(interval time.Duration) *Scheduler {
	return &Scheduler{
		interval: interval,
		done:     make(chan struct{}),
	}
}

func (s *Scheduler) SetFlush(fn func()) {
	s.flushFn = fn
}

func (s *Scheduler) Enqueue() {
	s.mu.Lock()
	s.dirty = true
	s.mu.Unlock()
}

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

func (s *Scheduler) Stop() {
	log.Println("[scheduler] stopping")
	close(s.done)
	if s.ticker != nil {
		s.ticker.Stop()
	}
}
