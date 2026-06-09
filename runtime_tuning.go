package fugo

import (
	"os"
	"runtime/debug"
	"strconv"

	"github.com/sazardev/fugo/flog"
)

// defaultGCPercent raises the GC trigger from Go's default of 100. The render
// loop re-walks the retained tree and re-marshals props every dirty frame, so
// it produces a steady stream of short-lived garbage. Collecting less often
// trades a modest amount of heap for fewer stop-the-world pauses inside the
// 16ms frame budget. It is only applied when the user has not set GOGC.
const defaultGCPercent = 200

// tuneRuntime applies render-loop-friendly GC settings. Both knobs are opt-out:
//
//   - FUGO_GOGC overrides the GC target percent. If unset and the standard GOGC
//     env is also unset, a tuned default (defaultGCPercent) is applied; if the
//     user set GOGC, their choice is left untouched.
//   - FUGO_GOMEMLIMIT sets a soft memory limit in bytes (the standard
//     GOMEMLIMIT env is already honored by the runtime on its own).
func tuneRuntime() {
	if v := os.Getenv("FUGO_GOGC"); v != "" {
		if pct, err := strconv.Atoi(v); err == nil {
			debug.SetGCPercent(pct)
			flog.Infof("GOGC set to %d (FUGO_GOGC)", pct)
		} else {
			flog.Errorf("invalid FUGO_GOGC=%q: %v", v, err)
		}
	} else if os.Getenv("GOGC") == "" {
		debug.SetGCPercent(defaultGCPercent)
		flog.Debugf("GOGC defaulted to %d (set FUGO_GOGC to override)", defaultGCPercent)
	}

	if v := os.Getenv("FUGO_GOMEMLIMIT"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			debug.SetMemoryLimit(n)
			flog.Infof("GOMEMLIMIT set to %d bytes (FUGO_GOMEMLIMIT)", n)
		} else {
			flog.Errorf("invalid FUGO_GOMEMLIMIT=%q (want positive byte count)", v)
		}
	}
}
