package fugo

import (
	"strconv"
	"testing"

	"github.com/sazardev/fugo/engine"
	"github.com/sazardev/fugo/fg"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

const clickEvent = "click"

// pumpEvents runs every task queued by HandleEvent/Mutate — the work the
// render goroutine would perform before its next flush.
func pumpEvents(app *App) {
	for _, fn := range app.scheduler.DrainTasks() {
		fn()
	}
}

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

	app.HandleEvent(&fugov1.ClientEvent{NodeId: strconv.FormatUint(uint64(id), 10), EventType: clickEvent})
	pumpEvents(app)

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

	app.HandleEvent(&fugov1.ClientEvent{NodeId: "9999", EventType: clickEvent})

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

// TestHandlersConcurrentAccess dispatches events while the handler registry is
// being re-collected (the scheduler's flush path), so `go test -race` catches
// any unguarded concurrent access to App.handlers.
func TestHandlersConcurrentAccess(t *testing.T) {
	app := NewApp(AppOptions{})

	btn := fg.Button("x").OnClick(func(_ fg.Event) {})
	_, m := fg.BuildTree(fg.Column(btn))
	app.collectHandlers(m)

	var id uint32
	for k := range app.handlers {
		id = k
	}

	ev := &fugov1.ClientEvent{NodeId: strconv.FormatUint(uint64(id), 10), EventType: clickEvent}

	const iterations = 2000

	done := make(chan struct{})
	go func() {
		for range iterations {
			app.collectHandlers(m) // writer: mimics flush re-collecting handlers
		}
		close(done)
	}()

	for range iterations {
		app.HandleEvent(ev) // reader: mimics the transport goroutine
	}

	<-done

	// Sanity: the handler is still resolvable after the concurrent storm.
	app.handlersMu.RLock()
	_, ok := app.handlers[id]
	app.handlersMu.RUnlock()

	if !ok {
		t.Fatalf("handler for node %d went missing after concurrent access", id)
	}
}

// TestClientDisconnectedDropsPendingReplies verifies a disconnect drops every
// pending host-service callback instead of leaking it forever.
func TestClientDisconnectedDropsPendingReplies(t *testing.T) {
	app, stream, ctx := newHostTestApp()

	called := false
	ctx.Clipboard().Read(func(string) {
		called = true
	})

	if host := stream.onlyHost(t); host.GetRequestId() == 0 {
		t.Fatal("expected a pending request id")
	}

	app.hostMu.Lock()
	pending := len(app.hostReqs)
	app.hostMu.Unlock()
	if pending != 1 {
		t.Fatalf("pending = %d, want 1 before disconnect", pending)
	}

	app.ClientDisconnected()

	app.hostMu.Lock()
	pending = len(app.hostReqs)
	app.hostMu.Unlock()
	if pending != 0 {
		t.Errorf("pending = %d after ClientDisconnected, want 0", pending)
	}
	if called {
		t.Error("dropped callback must not fire")
	}
}

// TestCollectHandlersPrunesStaleEntries verifies a fresh collection replaces —
// not accumulates — the registry, so widgets that left the tree stop being
// reachable and the map cannot grow without bound.
func TestCollectHandlersPrunesStaleEntries(t *testing.T) {
	app := NewApp(AppOptions{})

	btn := fg.Button("x").OnClick(func(_ fg.Event) {})
	_, m := fg.BuildTree(fg.Column(btn))
	app.collectHandlers(m)
	if len(app.handlers) != 1 {
		t.Fatalf("after first collect: %d handlers, want 1", len(app.handlers))
	}

	// Second frame: the button is gone; an unrelated handlerless tree remains.
	_, empty := fg.BuildTree(fg.Column(fg.Text("no handlers here")))
	app.collectHandlers(empty)

	if got := len(app.handlers); got != 0 {
		t.Errorf("stale handler survived re-collection: %d entries remain", got)
	}
}
