package engine

import (
	"log"
	"sync"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

type RenderStream interface {
	Send(payload *fugov1.RenderPayload) error
}

type Reconciler struct {
	stream  RenderStream
	pending []*fugov1.RenderPayload
	seq     uint64
	mu      sync.Mutex
}

func NewReconciler() *Reconciler {
	return &Reconciler{}
}

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

func (r *Reconciler) ClearStream() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stream = nil
}

func (r *Reconciler) SendFullTree(tree *fugov1.WidgetTree) {
	r.seq++
	payload := &fugov1.RenderPayload{
		Payload: &fugov1.RenderPayload_FullTree{FullTree: tree},
	}
	r.send(payload)
}

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
