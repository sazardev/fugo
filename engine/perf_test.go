package engine_test

import (
	"testing"
	"time"

	"github.com/sazardev/fugo/engine"
)

// TestDiffPerformanceBudget guards against gross algorithmic regressions: a
// 1000-node no-change diff must stay well under a frame. The fast path makes it
// ~25µs; the 1ms ceiling is race-detector-safe yet still catches an accidental
// O(n^2) blow-up or the loss of the no-change short-circuit's zero-alloc path.
func TestDiffPerformanceBudget(t *testing.T) {
	oldTree := makeTree(1000)
	newTree := makeTree(1000)

	const iterations = 300

	start := time.Now()
	for range iterations {
		engine.Diff(oldTree, newTree)
	}
	nsPerOp := time.Since(start).Nanoseconds() / iterations

	const budgetNs = 1_000_000 // 1ms
	if nsPerOp > budgetNs {
		t.Errorf("Diff(1000 nodes, no change) = %d ns/op, exceeds %d ns budget", nsPerOp, budgetNs)
	}
}
