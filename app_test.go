package fugo

import (
	"strconv"
	"testing"

	"github.com/sazardev/fugo/engine"
	"github.com/sazardev/fugo/fg"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

// TestHandleEventDispatch verifies a ClientEvent is routed to the handler of
// the widget whose node id matches.
func TestHandleEventDispatch(t *testing.T) {
	app := NewApp(AppOptions{})

	clicked := false
	btn := fg.Button("x").OnClick(func(_ fg.Event) { clicked = true })

	_, m := fg.BuildTree(fg.Column(btn))
	app.collectHandlers(m)

	if len(app.handlers) != 1 {
		t.Fatalf("expected 1 registered handler, got %d", len(app.handlers))
	}

	var id uint32
	for k := range app.handlers {
		id = k
	}

	app.HandleEvent(&fugov1.ClientEvent{NodeId: strconv.FormatUint(uint64(id), 10), EventType: "click"})

	if !clicked {
		t.Error("expected button handler to fire for matching node id")
	}
}

// TestHandleEventUnknownNode verifies an event for an unregistered node is a
// no-op (no panic, no handler fired).
func TestHandleEventUnknownNode(t *testing.T) {
	app := NewApp(AppOptions{})

	clicked := false
	btn := fg.Button("x").OnClick(func(_ fg.Event) { clicked = true })

	_, m := fg.BuildTree(fg.Column(btn))
	app.collectHandlers(m)

	app.HandleEvent(&fugov1.ClientEvent{NodeId: "9999", EventType: "click"})

	if clicked {
		t.Error("handler must not fire for an unknown node id")
	}
}

// TestUpdateCycleProducesPatch exercises the core render loop:
// build → mutate retained widget → rebuild-with-merge → diff yields a patch.
func TestUpdateCycleProducesPatch(t *testing.T) {
	txt := fg.Text("0")
	root := fg.Column(txt)

	oldTree, m := fg.BuildTree(root)

	txt.SetText("1")
	newTree, _ := fg.BuildTreeWithMerge(root, m)

	patches := engine.Diff(oldTree, newTree)
	if len(patches) == 0 {
		t.Error("expected at least one patch after mutating retained text widget")
	}
}

// TestUpdateCycleNoChangeNoPatch verifies an identical rebuild produces no
// patches (the diff short-circuits unchanged props).
func TestUpdateCycleNoChangeNoPatch(t *testing.T) {
	root := fg.Column(fg.Text("stable"))

	oldTree, m := fg.BuildTree(root)
	newTree, _ := fg.BuildTreeWithMerge(root, m)

	if patches := engine.Diff(oldTree, newTree); len(patches) != 0 {
		t.Errorf("expected no patches for an unchanged tree, got %d", len(patches))
	}
}
