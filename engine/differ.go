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
	oldMap map[uint32]*fugov1.WidgetNode
}

func Diff(oldTree, newTree *fugov1.WidgetTree) []Patch {
	if oldTree == nil {
		return fullCreate(newTree)
	}

	s := &diffState{
		oldMap: indexByID(oldTree.GetNodes()),
	}

	var patches []Patch

	for _, newNode := range newTree.GetNodes() {
		oldNode := s.oldMap[newNode.GetId()]

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

func fullCreate(tree *fugov1.WidgetTree) []Patch {
	var patches []Patch
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
