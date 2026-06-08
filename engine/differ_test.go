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

// applyPatches simulates the Flutter client: it seeds a flat id→node set from
// the old tree and replays patches in order, mirroring fugo_renderer.dart. It
// fails the test if any UPDATE/REORDER/DELETE targets an id the client does not
// hold — the invariant the diff must never violate. (A key-matched node whose
// depth-first id had shifted used to emit patches for ids the client never
// received, silently desyncing the tree.)
func applyPatches(t *testing.T, before *fugov1.WidgetTree, patches []engine.Patch) map[uint32]bool {
	t.Helper()

	client := idSet(before)

	for _, p := range patches {
		switch p.Op {
		case fugov1.PatchOp_PATCH_CREATE, fugov1.PatchOp_PATCH_REPLACE:
			client[p.NodeID] = true
		case fugov1.PatchOp_PATCH_UPDATE, fugov1.PatchOp_PATCH_REORDER:
			if !client[p.NodeID] {
				t.Fatalf("%v targets node %d the client does not hold", p.Op, p.NodeID)
			}
		case fugov1.PatchOp_PATCH_DELETE:
			if !client[p.NodeID] {
				t.Fatalf("DELETE targets node %d the client does not hold", p.NodeID)
			}

			delete(client, p.NodeID)
		}
	}

	return client
}

func idSet(tree *fugov1.WidgetTree) map[uint32]bool {
	s := make(map[uint32]bool, len(tree.GetNodes()))
	for _, n := range tree.GetNodes() {
		s[n.GetId()] = true
	}

	return s
}

// TestDiff_PatchesAlwaysApplicable is the regression guard for the removed
// keyed-match path: across reorders and inserts (where depth-first ids shift),
// every patch must reference an id the client can resolve, and replaying the
// patch stream must leave the client holding exactly the new tree's id set.
func TestDiff_PatchesAlwaysApplicable(t *testing.T) {
	cases := []struct {
		name          string
		before, after *fugov1.WidgetTree
	}{
		{
			name: "reorder with shifted ids",
			before: &fugov1.WidgetTree{Root: 1, Nodes: []*fugov1.WidgetNode{
				{Id: 1, Type: fugov1.WidgetType_COLUMN, Key: "list", Children: []uint32{2, 3}},
				{Id: 2, Type: fugov1.WidgetType_TEXT, Key: "item_a", Props: []byte("hello")},
				{Id: 3, Type: fugov1.WidgetType_TEXT, Key: "item_b", Props: []byte("world")},
			}},
			after: &fugov1.WidgetTree{Root: 5, Nodes: []*fugov1.WidgetNode{
				{Id: 5, Type: fugov1.WidgetType_COLUMN, Key: "list", Children: []uint32{6, 7}},
				{Id: 6, Type: fugov1.WidgetType_TEXT, Key: "item_b", Props: []byte("updated")},
				{Id: 7, Type: fugov1.WidgetType_TEXT, Key: "item_a", Props: []byte("hello")},
			}},
		},
		{
			name: "insert into list",
			before: &fugov1.WidgetTree{Root: 1, Nodes: []*fugov1.WidgetNode{
				{Id: 1, Type: fugov1.WidgetType_COLUMN, Children: []uint32{2, 3}},
				{Id: 2, Type: fugov1.WidgetType_TEXT, Props: []byte("a")},
				{Id: 3, Type: fugov1.WidgetType_TEXT, Props: []byte("b")},
			}},
			after: &fugov1.WidgetTree{Root: 1, Nodes: []*fugov1.WidgetNode{
				{Id: 1, Type: fugov1.WidgetType_COLUMN, Children: []uint32{2, 3, 4}},
				{Id: 2, Type: fugov1.WidgetType_TEXT, Props: []byte("a")},
				{Id: 3, Type: fugov1.WidgetType_TEXT, Props: []byte("x")},
				{Id: 4, Type: fugov1.WidgetType_TEXT, Props: []byte("b")},
			}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			patches := engine.Diff(tc.before, tc.after)
			client := applyPatches(t, tc.before, patches)

			want := idSet(tc.after)
			if len(client) != len(want) {
				t.Fatalf("after applying patches client holds %d nodes, want %d", len(client), len(want))
			}

			for id := range want {
				if !client[id] {
					t.Errorf("client missing new node %d after applying patches", id)
				}
			}
		})
	}
}
