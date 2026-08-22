//go:build race

package engine_test

// raceEnabled reports whether this test binary was built with the race
// detector (-race). Wall-clock assertions are meaningless under it — see
// TestDiffPerformanceBudget.
const raceEnabled = true
