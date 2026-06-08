package flog_test

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/sazardev/fugo/flog"
)

// TestLevelGating verifies that each level emits exactly the categories it
// should: Errorf always, Infof at Info+, Debugf only at Debug.
func TestLevelGating(t *testing.T) {
	prev := log.Default().Writer()
	flags := log.Flags()
	t.Cleanup(func() {
		log.SetOutput(prev)
		log.SetFlags(flags)
	})

	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFlags(0)

	cases := []struct {
		level     flog.Level
		wantDebug bool
		wantInfo  bool
	}{
		{flog.Silent, false, false},
		{flog.Info, false, true},
		{flog.Debug, true, true},
	}

	for _, tc := range cases {
		flog.SetLevel(tc.level)

		buf.Reset()
		flog.Debugf("dbg")
		if got := strings.Contains(buf.String(), "dbg"); got != tc.wantDebug {
			t.Errorf("level %d: Debugf emitted=%v, want %v", tc.level, got, tc.wantDebug)
		}

		buf.Reset()
		flog.Infof("nfo")
		if got := strings.Contains(buf.String(), "nfo"); got != tc.wantInfo {
			t.Errorf("level %d: Infof emitted=%v, want %v", tc.level, got, tc.wantInfo)
		}

		buf.Reset()
		flog.Errorf("err")
		if !strings.Contains(buf.String(), "err") {
			t.Errorf("level %d: Errorf must always emit", tc.level)
		}
	}
}
