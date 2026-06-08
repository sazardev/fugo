package engine_test

import (
	"testing"

	"github.com/sazardev/fugo/engine"
	fugov1 "github.com/sazardev/fugo/transport/proto/fugo/v1"
)

func TestDiff_FullCreate(t *testing.T) {
	newTree := &fugov1.WidgetTree{
		Root: 1,
		Nodes: []*fugov1.WidgetNode{
			{Id: 1, Type: fugov1.WidgetType_TEXT, Props: []byte("hello")},
		},
	}

	patches := engine.Diff(nil, newTree)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	if patches[0].Op != fugov1.PatchOp_PATCH_CREATE {
		t.Errorf("expected CREATE, got %v", patches[0].Op)
	}
}

func TestDiff_Update(t *testing.T) {
	oldTree := &fugov1.WidgetTree{
		Root: 1,
		Nodes: []*fugov1.WidgetNode{
			{Id: 1, Type: fugov1.WidgetType_TEXT, Props: []byte("old")},
		},
	}
	newTree := &fugov1.WidgetTree{
		Root: 1,
		Nodes: []*fugov1.WidgetNode{
			{Id: 1, Type: fugov1.WidgetType_TEXT, Props: []byte("new")},
		},
	}

	patches := engine.Diff(oldTree, newTree)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	if patches[0].Op != fugov1.PatchOp_PATCH_UPDATE {
		t.Errorf("expected UPDATE, got %v", patches[0].Op)
	}
}

func TestDiff_Delete(t *testing.T) {
	oldTree := &fugov1.WidgetTree{
		Root: 1,
		Nodes: []*fugov1.WidgetNode{
			{Id: 1, Type: fugov1.WidgetType_TEXT, Props: []byte("x")},
			{Id: 2, Type: fugov1.WidgetType_TEXT, Props: []byte("y")},
		},
	}
	newTree := &fugov1.WidgetTree{
		Root: 1,
		Nodes: []*fugov1.WidgetNode{
			{Id: 1, Type: fugov1.WidgetType_TEXT, Props: []byte("x")},
		},
	}

	patches := engine.Diff(oldTree, newTree)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	if patches[0].Op != fugov1.PatchOp_PATCH_DELETE {
		t.Errorf("expected DELETE, got %v", patches[0].Op)
	}

	if patches[0].NodeID != 2 {
		t.Errorf("expected node 2 deleted, got %d", patches[0].NodeID)
	}
}

func TestDiff_Replace(t *testing.T) {
	oldTree := &fugov1.WidgetTree{
		Root: 1,
		Nodes: []*fugov1.WidgetNode{
			{Id: 1, Type: fugov1.WidgetType_TEXT, Props: []byte("x")},
		},
	}
	newTree := &fugov1.WidgetTree{
		Root: 1,
		Nodes: []*fugov1.WidgetNode{
			{Id: 1, Type: fugov1.WidgetType_BUTTON, Props: []byte("y")},
		},
	}

	patches := engine.Diff(oldTree, newTree)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	if patches[0].Op != fugov1.PatchOp_PATCH_REPLACE {
		t.Errorf("expected REPLACE, got %v", patches[0].Op)
	}
}

func TestDiff_Reorder(t *testing.T) {
	oldTree := &fugov1.WidgetTree{
		Root: 1,
		Nodes: []*fugov1.WidgetNode{
			{Id: 1, Type: fugov1.WidgetType_COLUMN, Children: []uint32{2, 3}},
			{Id: 2, Type: fugov1.WidgetType_TEXT, Props: []byte("a")},
			{Id: 3, Type: fugov1.WidgetType_TEXT, Props: []byte("b")},
		},
	}
	newTree := &fugov1.WidgetTree{
		Root: 1,
		Nodes: []*fugov1.WidgetNode{
			{Id: 1, Type: fugov1.WidgetType_COLUMN, Children: []uint32{3, 2}},
			{Id: 2, Type: fugov1.WidgetType_TEXT, Props: []byte("a")},
			{Id: 3, Type: fugov1.WidgetType_TEXT, Props: []byte("b")},
		},
	}

	patches := engine.Diff(oldTree, newTree)
	if len(patches) != 1 {
		t.Fatalf("expected 1 patch, got %d", len(patches))
	}

	if patches[0].Op != fugov1.PatchOp_PATCH_REORDER {
		t.Errorf("expected REORDER, got %v", patches[0].Op)
	}
}

func TestDiff_NoChange(t *testing.T) {
	tree := &fugov1.WidgetTree{
		Root: 1,
		Nodes: []*fugov1.WidgetNode{
			{Id: 1, Type: fugov1.WidgetType_TEXT, Props: []byte("x")},
		},
	}

	patches := engine.Diff(tree, tree)
	if len(patches) != 0 {
		t.Fatalf("expected 0 patches, got %d", len(patches))
	}
}

func TestDiff_CreateDeleteUpdate(t *testing.T) {
	oldTree := &fugov1.WidgetTree{
		Root: 1,
		Nodes: []*fugov1.WidgetNode{
			{Id: 1, Type: fugov1.WidgetType_COLUMN, Children: []uint32{2, 3}},
			{Id: 2, Type: fugov1.WidgetType_TEXT, Props: []byte("a")},
			{Id: 3, Type: fugov1.WidgetType_TEXT, Props: []byte("b")},
		},
	}
	newTree := &fugov1.WidgetTree{
		Root: 1,
		Nodes: []*fugov1.WidgetNode{
			{Id: 1, Type: fugov1.WidgetType_COLUMN, Children: []uint32{2, 4}},
			{Id: 2, Type: fugov1.WidgetType_TEXT, Props: []byte("updated")},
			{Id: 4, Type: fugov1.WidgetType_BUTTON, Props: []byte("click")},
		},
	}

	patches := engine.Diff(oldTree, newTree)
	ops := make(map[fugov1.PatchOp]bool)

	for _, p := range patches {
		ops[p.Op] = true
	}

	if !ops[fugov1.PatchOp_PATCH_UPDATE] {
		t.Error("expected UPDATE for node 2")
	}

	if !ops[fugov1.PatchOp_PATCH_DELETE] {
		t.Error("expected DELETE for node 3")
	}

	if !ops[fugov1.PatchOp_PATCH_CREATE] {
		t.Error("expected CREATE for node 4")
	}

	if !ops[fugov1.PatchOp_PATCH_REORDER] {
		t.Error("expected REORDER for node 1")
	}
}

func TestDiff_KeyBasedMatch(t *testing.T) {
	oldTree := &fugov1.WidgetTree{
		Root: 1,
		Nodes: []*fugov1.WidgetNode{
			{Id: 1, Type: fugov1.WidgetType_COLUMN, Key: "list", Children: []uint32{2, 3}},
			{Id: 2, Type: fugov1.WidgetType_TEXT, Key: "item_a", Props: []byte("hello")},
			{Id: 3, Type: fugov1.WidgetType_TEXT, Key: "item_b", Props: []byte("world")},
		},
	}

	newTree := &fugov1.WidgetTree{
		Root: 5,
		Nodes: []*fugov1.WidgetNode{
			{Id: 5, Type: fugov1.WidgetType_COLUMN, Key: "list", Children: []uint32{6, 7}},
			{Id: 6, Type: fugov1.WidgetType_TEXT, Key: "item_b", Props: []byte("updated")},
			{Id: 7, Type: fugov1.WidgetType_TEXT, Key: "item_a", Props: []byte("hello")},
		},
	}

	patches := engine.Diff(oldTree, newTree)

	for _, p := range patches {
		switch p.Op {
		case fugov1.PatchOp_PATCH_DELETE:
			t.Errorf("unexpected DELETE for node %d (key-based match should prevent it)", p.NodeID)
		case fugov1.PatchOp_PATCH_CREATE:
			t.Errorf("unexpected CREATE (key-based match should detect existing keys)")
		case fugov1.PatchOp_PATCH_REPLACE:
			t.Errorf("unexpected REPLACE (types haven't changed)")
		case fugov1.PatchOp_PATCH_UPDATE, fugov1.PatchOp_PATCH_REORDER:
			// expected for a keyed update + reorder; asserted below
		}
	}

	foundUpdate := false
	foundReorder := false

	for _, p := range patches {
		if p.Op == fugov1.PatchOp_PATCH_UPDATE {
			foundUpdate = true
		}

		if p.Op == fugov1.PatchOp_PATCH_REORDER {
			foundReorder = true
		}
	}

	if !foundUpdate {
		t.Error("expected UPDATE for key-matched node with changed props")
	}

	if !foundReorder {
		t.Error("expected REORDER for parent whose children changed order")
	}
}
