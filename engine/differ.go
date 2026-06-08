package engine

import (
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

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

func Diff(oldTree, newTree *fugov1.WidgetTree) []Patch {
	if oldTree == nil {
		return fullCreate(newTree)
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
