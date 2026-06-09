package engine

import (
	"sync"

	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

// oldMapPool recycles the id→node lookup map allocated on the diff's change
// path. The no-change fast path returns before touching it, so the zero-alloc
// guarantee (TestDiffNoChangeZeroAlloc) is unaffected; this only trims the
// per-frame allocation when something actually changed.
var oldMapPool = sync.Pool{
	New: func() any { return make(map[uint32]*fugov1.WidgetNode) },
}

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

// Diff compares the previous and next widget trees and returns the minimal set
// of patches (create/update/delete/replace/reorder) needed to bring the client
// in sync. Nodes are matched by ID — the retained tree assigns ids
// deterministically depth-first every frame, so identity is positional. When
// oldTree is nil it emits a full create for every node in newTree.
//
// Every emitted patch references a node id the client either already holds
// (from the previous tree) or is creating in the same batch, so the client —
// which keys nodes by id — can always apply it. (Cross-frame matching by Key
// was removed: when a keyed node's depth-first id shifted, it produced
// UPDATE/REORDER patches for ids the client had never received, silently
// desyncing the tree. A reorder/insert now emits applicable CREATE/UPDATE/
// DELETE/REORDER patches instead.)
func Diff(oldTree, newTree *fugov1.WidgetTree) []Patch {
	if oldTree == nil {
		return fullCreate(newTree)
	}

	// Fast path: the retained tree assigns ids in a stable order every frame,
	// so an allocation-free positional compare lets the common "nothing changed"
	// case return early without building the lookup map.
	if treesEqual(oldTree, newTree) {
		return nil
	}

	oldMap := oldMapPool.Get().(map[uint32]*fugov1.WidgetNode)
	for _, n := range oldTree.GetNodes() {
		oldMap[n.GetId()] = n
	}

	var patches []Patch

	for _, newNode := range newTree.GetNodes() {
		oldNode := oldMap[newNode.GetId()]

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

		delete(oldMap, newNode.GetId())
	}

	for id := range oldMap {
		patches = append(patches, Patch{
			Op:     fugov1.PatchOp_PATCH_DELETE,
			NodeID: id,
		})
	}

	// Clear before returning to the pool so it stops pinning the previous
	// frame's nodes; a map is a reference type, so Put itself does not allocate.
	clear(oldMap)
	oldMapPool.Put(oldMap)

	return patches
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
