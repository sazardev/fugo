package engine

import (
	"log"
	"sync"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

// RenderStream is the outbound channel to the client. It is satisfied by the
// gRPC render stream adapter and lets the Reconciler push render payloads
// without depending on the transport package.
type RenderStream interface {
	Send(payload *fugov1.RenderPayload) error
}

// Reconciler serializes render output to the client. It holds the current
// RenderStream and buffers payloads while no client is connected, replaying
// them once a stream is attached. All methods are safe for concurrent use.
type Reconciler struct {
	stream  RenderStream
	pending []*fugov1.RenderPayload
	seq     uint64
	mu      sync.Mutex
}

// NewReconciler returns a Reconciler with no stream attached; payloads sent
// before a stream is set are buffered until SetStream is called.
func NewReconciler() *Reconciler {
	return &Reconciler{}
}

// SetStream attaches the given stream and immediately flushes any payloads that
// were buffered while no client was connected.
func (r *Reconciler) SetStream(stream RenderStream) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.stream = stream
	for _, p := range r.pending {
		if err := stream.Send(p); err != nil {
			log.Printf("[engine] pending send error: %v", err)
		}
	}

	r.pending = nil
}

// ClearStream detaches the current stream, typically after the client
// disconnects; subsequent payloads are buffered until a new stream is set.
func (r *Reconciler) ClearStream() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stream = nil
}

// SendFullTree sends the entire widget tree to the client, used for the initial
// render or a full resync rather than an incremental patch set.
func (r *Reconciler) SendFullTree(tree *fugov1.WidgetTree) {
	r.seq++
	payload := &fugov1.RenderPayload{
		Payload: &fugov1.RenderPayload_FullTree{FullTree: tree},
	}
	r.send(payload)
}

// SendPatches converts the diff patches into their protobuf form, tags them
// with a monotonically increasing sequence number, and sends them to the client.
func (r *Reconciler) SendPatches(patches []Patch) {
	r.seq++

	pbPatches := make([]*fugov1.Patch, len(patches))
	for i, p := range patches {
		pbPatches[i] = &fugov1.Patch{
			Op:       p.Op,
			NodeId:   p.NodeID,
			Node:     p.Node,
			Props:    p.Props,
			Children: p.Children,
			ParentId: p.ParentID,
		}
	}

	payload := &fugov1.RenderPayload{
		Payload: &fugov1.RenderPayload_Patches{
			Patches: &fugov1.PatchList{
				Patches: pbPatches,
				SeqNum:  r.seq,
			},
		},
	}
	r.send(payload)
}

// SendWindowCommand sends an out-of-band window-control command (set title or
// size, minimize, maximize, center, fullscreen) to the client.
func (r *Reconciler) SendWindowCommand(cmd *fugov1.WindowCommand) {
	r.send(&fugov1.RenderPayload{
		Payload: &fugov1.RenderPayload_Window{Window: cmd},
	})
}

// SendHostCommand sends an out-of-band host-service request (clipboard access,
// native file dialog) to the client. Requests that expect a reply carry a
// non-zero RequestId and the client answers with a "host" ClientEvent.
func (r *Reconciler) SendHostCommand(cmd *fugov1.HostCommand) {
	r.send(&fugov1.RenderPayload{
		Payload: &fugov1.RenderPayload_Host{Host: cmd},
	})
}

func (r *Reconciler) send(payload *fugov1.RenderPayload) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.stream != nil {
		if err := r.stream.Send(payload); err != nil {
			log.Printf("[engine] send error: %v", err)
		}
	} else {
		r.pending = append(r.pending, payload)
	}
}
