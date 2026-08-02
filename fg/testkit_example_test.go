package fg_test

import (
	"testing"

	"github.com/sazardev/fugo/fg"
)

// TestCheckboxTogglesCompanionText is the canonical example of testing a Fugo
// UI without gRPC or Flutter: build the tree once with fg.BuildTree, grab a
// concrete widget reference, drive it with a synthetic event from the
// testkit, and assert on exported fields — mirroring the Checkbox +
// companion-Text pattern used in the CLI's showcase template
// (cmd/fugo/templates.go: cbStatus/cb).
func TestCheckboxTogglesCompanionText(t *testing.T) {
	status := fg.Text("off")
	cb := fg.Checkbox("Enable feature").OnChange(func(e fg.Event) {
		if string(e.Data) == "1" {
			status.SetText("on")
		} else {
			status.SetText("off")
		}
	})

	root := fg.Column(cb, status)

	// BuildTree assigns ids; the returned map lets callers look widgets up by
	// node id, but here we already hold direct references (cb, status), which
	// is the common case in a test that built the tree itself.
	_, _ = fg.BuildTree(root)

	if status.Value != "off" {
		t.Fatalf("initial status = %q, want %q", status.Value, "off")
	}

	cb.Handle(fg.BoolEvent(true))

	if !cb.HasHandler() {
		t.Fatal("expected checkbox to have a registered OnChange handler")
	}

	if status.Value != "on" {
		t.Fatalf("after BoolEvent(true), status = %q, want %q", status.Value, "on")
	}

	cb.Handle(fg.BoolEvent(false))

	if status.Value != "off" {
		t.Fatalf("after BoolEvent(false), status = %q, want %q", status.Value, "off")
	}
}
