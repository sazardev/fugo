package engine_test

import (
	"sync"
	"testing"
	"time"

	"github.com/sazardev/fugo/engine"
)

func TestScheduler_Coalesces(t *testing.T) {
	var mu sync.Mutex
	var callCount int

	sched := engine.NewScheduler(10 * time.Millisecond)
	sched.SetFlush(func() {
		mu.Lock()
		callCount++
		mu.Unlock()
	})
	sched.Start()
	defer sched.Stop()

	for range 10 {
		sched.Enqueue()
		time.Sleep(1 * time.Millisecond)
	}

	time.Sleep(30 * time.Millisecond)

	mu.Lock()
	count := callCount
	mu.Unlock()

	if count < 1 {
		t.Errorf("expected at least 1 flush, got %d", count)
	}
	if count > 3 {
		t.Errorf("expected at most 3 flushes in 40ms, got %d (coalescing failed)", count)
	}
}

func TestScheduler_FlushOnlyWhenDirty(t *testing.T) {
	var mu sync.Mutex
	var callCount int

	sched := engine.NewScheduler(5 * time.Millisecond)
	sched.SetFlush(func() {
		mu.Lock()
		callCount++
		mu.Unlock()
	})
	sched.Start()
	defer sched.Stop()

	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	count := callCount
	mu.Unlock()

	if count != 0 {
		t.Errorf("expected 0 flushes without enqueue, got %d", count)
	}
}

func TestScheduler_Stop(_ *testing.T) {
	sched := engine.NewScheduler(5 * time.Millisecond)
	sched.Start()
	sched.Stop()

	sched.Enqueue()
	time.Sleep(20 * time.Millisecond)
}
