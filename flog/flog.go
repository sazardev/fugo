// Package flog is Fugo's tiny leveled logger for runtime ("[fugo]") messages.
//
// The level is taken from the FUGO_LOG environment variable so the CLI can
// drive it via --quiet/--verbose without the app importing the CLI:
//
//	silent | off | quiet   → only Errorf
//	info (default)         → Errorf + Infof (lifecycle)
//	debug | verbose        → everything, including per-frame/per-event Debugf
//
// It is a leaf package (imports only the standard library) so every other
// package — app, transport, supervisor, the demo — can use it without import
// cycles.
package flog

import (
	"log"
	"os"
	"strings"
	"sync"
)

// Level controls how much the runtime logs.
type Level int

const (
	// Silent logs only errors.
	Silent Level = iota
	// Info logs errors and lifecycle messages (the default).
	Info
	// Debug logs everything, including high-frequency diagnostics.
	Debug
)

var (
	mu          sync.RWMutex
	level       Level
	initialized bool
)

func get() Level {
	mu.RLock()
	if initialized {
		l := level
		mu.RUnlock()

		return l
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()
	if !initialized {
		level = parse(os.Getenv("FUGO_LOG"))
		initialized = true
	}

	return level
}

func parse(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "silent", "off", "quiet", "none", "0":
		return Silent
	case "debug", "trace", "verbose", "2":
		return Debug
	default:
		return Info
	}
}

// SetLevel sets the level explicitly, overriding FUGO_LOG. Mainly for tests.
func SetLevel(l Level) {
	mu.Lock()
	defer mu.Unlock()
	level = l
	initialized = true
}

// Infof logs a lifecycle message (server start, client connected, window
// closed). Suppressed when the level is Silent.
func Infof(format string, a ...any) {
	if get() >= Info {
		log.Printf("[fugo] "+format, a...)
	}
}

// Debugf logs a high-frequency diagnostic (per-frame flush, per-event
// dispatch). Shown only at Debug level.
func Debugf(format string, a ...any) {
	if get() >= Debug {
		log.Printf("[fugo] "+format, a...)
	}
}

// Errorf logs an error or rejected request; shown at every level.
func Errorf(format string, a ...any) {
	log.Printf("[fugo] "+format, a...)
}
