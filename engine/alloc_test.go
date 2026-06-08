//go:build !race

package engine_test

import (
	"testing"

	"github.com/sazardev/fugo/engine"
)

// TestDiffNoChangeZeroAlloc locks in the no-change fast path's zero-allocation
// behavior. Unlike a wall-clock budget it is deterministic (it counts
// allocations, not nanoseconds), so it is not flaky on shared CI runners and
// catches the loss of the positional short-circuit the moment it falls back to
// building the lookup map. It is skipped under -race because the race detector
// instruments allocations; the CI bench job runs it without -race.
func TestDiffNoChangeZeroAlloc(t *testing.T) {
	oldTree := makeTree(1000)
	newTree := makeTree(1000)

	if avg := testing.AllocsPerRun(100, func() {
		engine.Diff(oldTree, newTree)
	}); avg != 0 {
		t.Errorf("Diff(1000 nodes, no change) = %.0f allocs/op, want 0 — the zero-alloc fast path regressed", avg)
	}
}
