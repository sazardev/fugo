package engine

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

// Patch describes a single mutation to apply to the client's widget tree:
// creating, updating, deleting, replacing, or reordering a node. The fields
// that are populated depend on Op (e.g. Props for an update, Children for a
// reorder, Node for a create/replace).
type Patch struct {
	Node     *fugov1.WidgetNode
	Props    []byte
	Children []uint32
	Op       fugov1.PatchOp
	NodeID   uint32
	ParentID uint32
}

type diffState struct {
	oldMap   map[uint32]*fugov1.WidgetNode
	oldKeyed map[string]*fugov1.WidgetNode
}

// Diff compares the previous and next widget trees and returns the minimal set
// of patches (create/update/delete/replace/reorder) needed to bring the client
// in sync. Nodes are matched by ID, falling back to their key when present;
// when oldTree is nil it emits a full create for every node in newTree.
func Diff(oldTree, newTree *fugov1.WidgetTree) []Patch {
	if oldTree == nil {
		return fullCreate(newTree)
	}

	// Fast path: the retained tree assigns ids in a stable order every frame,
	// so an allocation-free positional compare lets the common "nothing changed"
	// case return early without building the lookup maps.
	if treesEqual(oldTree, newTree) {
		return nil
	}

	s := &diffState{
		oldMap: indexByID(oldTree.GetNodes()),
	}

	s.oldKeyed = indexByKey(oldTree.GetNodes())

	var patches []Patch

	for _, newNode := range newTree.GetNodes() {
		oldNode := s.lookup(newNode)

		switch {
		case oldNode == nil:
			patches = append(patches, Patch{
				Op:     fugov1.PatchOp_PATCH_CREATE,
				NodeID: newNode.GetId(),
				Node:   newNode,
			})

		case oldNode.GetType() != newNode.GetType():
			patches = append(patches, Patch{
				Op:     fugov1.PatchOp_PATCH_REPLACE,
				NodeID: newNode.GetId(),
				Node:   newNode,
			})

		default:
			if !bytesEqual(oldNode.GetProps(), newNode.GetProps()) {
				patches = append(patches, Patch{
					Op:     fugov1.PatchOp_PATCH_UPDATE,
					NodeID: newNode.GetId(),
					Props:  newNode.GetProps(),
				})
			}

			if !uint32SliceEqual(oldNode.GetChildren(), newNode.GetChildren()) {
				patches = append(patches, Patch{
					Op:       fugov1.PatchOp_PATCH_REORDER,
					NodeID:   newNode.GetId(),
					Children: newNode.GetChildren(),
				})
			}
		}

		if oldNode != nil {
			delete(s.oldMap, oldNode.GetId())
			if oldNode.GetKey() != "" {
				delete(s.oldKeyed, oldNode.GetKey())
			}
		}

		delete(s.oldMap, newNode.GetId())
	}

	for id := range s.oldMap {
		patches = append(patches, Patch{
			Op:     fugov1.PatchOp_PATCH_DELETE,
			NodeID: id,
		})
	}

	return patches
}

func (s *diffState) lookup(newNode *fugov1.WidgetNode) *fugov1.WidgetNode {
	if old, ok := s.oldMap[newNode.GetId()]; ok {
		return old
	}

	if key := newNode.GetKey(); key != "" {
		return s.oldKeyed[key]
	}

	return nil
}

func fullCreate(tree *fugov1.WidgetTree) []Patch {
	patches := make([]Patch, 0, len(tree.GetNodes()))
	for _, node := range tree.GetNodes() {
		patches = append(patches, Patch{
			Op:     fugov1.PatchOp_PATCH_CREATE,
			NodeID: node.GetId(),
			Node:   node,
		})
	}

	return patches
}

// treesEqual reports whether a and b are structurally identical, comparing
// nodes positionally with no allocations. It is the diff's no-change fast path.
func treesEqual(a, b *fugov1.WidgetTree) bool {
	an, bn := a.GetNodes(), b.GetNodes()
	if len(an) != len(bn) {
		return false
	}

	for i := range an {
		x, y := an[i], bn[i]
		if x.GetId() != y.GetId() || x.GetType() != y.GetType() || x.GetKey() != y.GetKey() {
			return false
		}

		if !bytesEqual(x.GetProps(), y.GetProps()) {
			return false
		}

		if !uint32SliceEqual(x.GetChildren(), y.GetChildren()) {
			return false
		}
	}

	return true
}

func indexByID(nodes []*fugov1.WidgetNode) map[uint32]*fugov1.WidgetNode {
	m := make(map[uint32]*fugov1.WidgetNode, len(nodes))
	for _, n := range nodes {
		m[n.GetId()] = n
	}

	return m
}

func indexByKey(nodes []*fugov1.WidgetNode) map[string]*fugov1.WidgetNode {
	m := make(map[string]*fugov1.WidgetNode)
	for _, n := range nodes {
		if key := n.GetKey(); key != "" {
			m[key] = n
		}
	}

	return m
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func uint32SliceEqual(a, b []uint32) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
