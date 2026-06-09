package fugo

import (
	"strconv"
	"testing"

	"github.com/sazardev/fugo/engine"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

// captureStream records every payload the reconciler sends, so tests can
// inspect the host command (and its request id) that a Context call emitted.
type captureStream struct {
	payloads []*fugov1.RenderPayload
}

func (c *captureStream) Send(p *fugov1.RenderPayload) error {
	c.payloads = append(c.payloads, p)

	return nil
}

func newHostTestApp() (*App, *captureStream, *Context) {
	app := NewApp(AppOptions{})
	stream := &captureStream{}
	app.reconciler = engine.NewReconciler()
	app.reconciler.SetStream(stream)

	return app, stream, &Context{app: app}
}

func (c *captureStream) onlyHost(t *testing.T) *fugov1.HostCommand {
	t.Helper()

	if len(c.payloads) != 1 {
		t.Fatalf("expected exactly 1 payload, got %d", len(c.payloads))
	}

	host := c.payloads[0].GetHost()
	if host == nil {
		t.Fatalf("expected a host command payload, got %v", c.payloads[0])
	}

	return host
}

// TestClipboardReadRoundTrip verifies a read emits a correlated request and the
// matching "host" reply invokes the callback with the result, then clears it.
func TestClipboardReadRoundTrip(t *testing.T) {
	app, stream, ctx := newHostTestApp()

	var got string
	called := false
	ctx.Clipboard().Read(func(text string) {
		got = text
		called = true
	})

	host := stream.onlyHost(t)
	if host.GetOp() != fugov1.HostOp_HOST_CLIPBOARD_READ {
		t.Errorf("op = %v, want HOST_CLIPBOARD_READ", host.GetOp())
	}
	if host.GetRequestId() == 0 {
		t.Fatal("a read must carry a non-zero request id")
	}

	app.HandleEvent(&fugov1.ClientEvent{
		NodeId:    strconv.FormatUint(host.GetRequestId(), 10),
		EventType: hostEventType,
		EventData: []byte("hello clip"),
	})

	if !called {
		t.Fatal("read callback was not invoked")
	}
	if got != "hello clip" {
		t.Errorf("got %q, want %q", got, "hello clip")
	}

	app.hostMu.Lock()
	pending := len(app.hostReqs)
	app.hostMu.Unlock()
	if pending != 0 {
		t.Errorf("pending host requests = %d, want 0 after reply", pending)
	}
}

// TestClipboardWriteFireAndForget verifies a write sends text with no request
// id and registers no pending callback.
func TestClipboardWriteFireAndForget(t *testing.T) {
	app, stream, ctx := newHostTestApp()

	ctx.Clipboard().Write("data")

	host := stream.onlyHost(t)
	if host.GetOp() != fugov1.HostOp_HOST_CLIPBOARD_WRITE {
		t.Errorf("op = %v, want HOST_CLIPBOARD_WRITE", host.GetOp())
	}
	if host.GetText() != "data" {
		t.Errorf("text = %q, want %q", host.GetText(), "data")
	}
	if host.GetRequestId() != 0 {
		t.Errorf("write must be fire-and-forget (request id 0), got %d", host.GetRequestId())
	}

	app.hostMu.Lock()
	pending := len(app.hostReqs)
	app.hostMu.Unlock()
	if pending != 0 {
		t.Errorf("write must not register a pending request, got %d", pending)
	}
}

// TestFileOpenRoundTrip verifies the open dialog forwards title + extensions and
// the reply path delivers the chosen path.
func TestFileOpenRoundTrip(t *testing.T) {
	app, stream, ctx := newHostTestApp()

	var got string
	ctx.Files().Open(FileDialog{Title: "Pick", Extensions: []string{"png", "jpg"}}, func(p string) {
		got = p
	})

	host := stream.onlyHost(t)
	if host.GetOp() != fugov1.HostOp_HOST_FILE_OPEN {
		t.Errorf("op = %v, want HOST_FILE_OPEN", host.GetOp())
	}
	if host.GetText() != "Pick" {
		t.Errorf("title = %q, want %q", host.GetText(), "Pick")
	}
	if exts := host.GetExtensions(); len(exts) != 2 || exts[0] != "png" || exts[1] != "jpg" {
		t.Errorf("extensions = %v, want [png jpg]", exts)
	}

	const chosen = `C:\tmp\a.png`
	app.HandleEvent(&fugov1.ClientEvent{
		NodeId:    strconv.FormatUint(host.GetRequestId(), 10),
		EventType: hostEventType,
		EventData: []byte(chosen),
	})

	if got != chosen {
		t.Errorf("got %q, want %q", got, chosen)
	}
}

// TestHostReplyUnknownRequestIsNoop verifies a reply for an unknown request id
// neither panics nor fires anything.
func TestHostReplyUnknownRequestIsNoop(_ *testing.T) {
	app, _, _ := newHostTestApp()

	app.HandleEvent(&fugov1.ClientEvent{
		NodeId:    "424242",
		EventType: hostEventType,
		EventData: []byte("orphan"),
	})
}
