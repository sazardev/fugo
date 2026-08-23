package engine

import (
	"testing"
)

// TestEnqueueTaskRunsBeforeFlush verifies the mutation-task contract: queued
// tasks execute on demand (DrainTasks) before the flush they triggered, and a
// burst of tasks coalesces into a single flush.
func TestEnqueueTaskRunsBeforeFlush(t *testing.T) {
	s := NewScheduler(0)
	flushes := 0
	s.SetFlush(func() { flushes++ })

	var order []string
	s.EnqueueTask(func() { order = append(order, "task1") })
	s.EnqueueTask(func() { order = append(order, "task2") })

	// Simulate one render-goroutine cycle: drain, then flush.
	tasks := s.DrainTasks()
	if len(tasks) != 2 {
		t.Fatalf("queued %d tasks, want 2", len(tasks))
	}
	for _, fn := range tasks {
		fn()
	}
	s.MaybeFlushForTest()

	if len(order) != 2 || order[0] != "task1" || order[1] != "task2" {
		t.Errorf("tasks ran out of order or were lost: %v", order)
	}
	if flushes != 1 {
		t.Errorf("flush ran %d times for one dirty cycle, want 1", flushes)
	}

	// Draining with nothing pending yields no tasks and no flush.
	if got := s.DrainTasks(); len(got) != 0 {
		t.Errorf("expected empty drain, got %d tasks", len(got))
	}
	s.MaybeFlushForTest()
	if flushes != 1 {
		t.Errorf("idle cycle flushed: %d total flushes, want 1", flushes)
	}
}
